//go:build js

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"log"
	"path/filepath"
	"syscall/js"
	"time"

	"github.com/richardlehane/siegfried"
	"github.com/richardlehane/siegfried/internal/checksum"
	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/decompress"
	"github.com/richardlehane/siegfried/pkg/pronom"

	//"github.com/richardlehane/siegfried/pkg/sets"
	"github.com/richardlehane/siegfried/pkg/static"
	"github.com/richardlehane/siegfried/pkg/writer"
	//"github.com/richardlehane/siegfried/pkg/pronom"
)

type output int

const (
	jsonOut output = iota
	yamlOut
	csvOut
	droidOut
)

func opts(args []js.Value) (output, checksum.HashTyp, bool) {

	fmt.Println("args: ", args)

	var out output
	var ht checksum.HashTyp = -1
	var z bool
	for _, v := range args {
		vv := v.String()
		switch vv {
		case "yaml":
			out = yamlOut
		case "csv":
			out = csvOut
		case "droid":
			out = droidOut
		case "z":
			z = true
			config.SetArchiveFilterPermissive(config.ListAllArcTypes())
		default:
			htyp := checksum.GetHash(v.String())
			if htyp > -1 {
				ht = htyp
			}
		}
	}
	return out, ht, z
}

func identifyRdr(
	s *siegfried.Siegfried,
	r io.Reader,
	w writer.Writer,
	path string,
	mime string,
	sz int64,
	mod time.Time,
	h hash.Hash,
	z bool,
	do bool,
) {
	b, berr := s.Buffer(r)
	defer s.Put(b)
	ids, err := s.IdentifyBuffer(b, berr, path, mime)
	if ids == nil {
		w.File(path, sz, mod.Format(time.RFC3339), nil, err, ids)
		return
	}
	// calculate checksum
	var cs []byte
	if h != nil {
		var i int64
		l := h.BlockSize()
		for ; ; i += int64(l) {
			buf, _ := b.Slice(i, l)
			if buf == nil {
				break
			}
			h.Write(buf)
		}
		cs = h.Sum(nil)
	}
	// decompress if an archive format
	if !z {
		w.File(path, sz, mod.Format(time.RFC3339), cs, err, ids)
		return
	}
	arc := decompress.IsArc(ids)
	if arc == config.None {
		w.File(path, sz, mod.Format(time.RFC3339), cs, err, ids)
		return
	}
	d, err := decompress.New(arc, b, path)
	if err != nil {
		w.File(path, sz, mod.Format(time.RFC3339), cs, fmt.Errorf("failed to decompress, got: %v", err), ids)
		return
	}
	// write the result
	w.File(path, sz, mod.Format(time.RFC3339), cs, err, ids)
	// decompress and recurse
	for err = d.Next(); ; err = d.Next() {
		if err != nil {
			if err == io.EOF {
				return
			}
			w.File(decompress.Arcpath(path, ""), 0, time.Time{}.Format(time.RFC3339), nil, fmt.Errorf("error occurred during decompression: %v", err), nil)
			return
		}
		if do {
			for _, v := range d.Dirs() {
				w.File(v, -1, "", nil, nil, nil)
			}
		}
		identifyRdr(s, d.Reader(), w, d.Path(), d.MIME(), d.Size(), d.Mod(), h, z, do)
	}
}

func identifyFile(
	s *siegfried.Siegfried,
	r *reader,
	f js.Value,
	w writer.Writer,
	h hash.Hash,
	z bool,
	do bool,
	dirs []string,
) error {
	var (
		name string
		mod  time.Time
	)
	promise := f.Call("getFile")
	val, err := await(promise)
	if err != nil {
		return err
	}
	name = filepath.Join(append(dirs, val.Get("name").String())...)
	modUnix := int64(val.Get("lastModified").Int())
	mod = time.UnixMilli(modUnix)
	r.reset(val)
	fmt.Println("identifysize: ", len(r.buf))
	identifyRdr(s, r, w, name, "", r.Size(), mod, h, z, do)
	return nil
}

func identifyFiles(
	s *siegfried.Siegfried,
	r *reader,
	fsh js.Value,
	w writer.Writer,
	h hash.Hash,
	z bool,
	do bool,
	dirs []string,
) error {
	// identify files is called by default and ineeded it calles
	// for the identification of a single file rather than a slice
	// of files...

	// kind seems to come from the FileSystemHandle Web API.
	// https://developer.mozilla.org/en-US/docs/Web/API/FileSystemHandle/kind
	//
	kind := fsh.Get("kind").String()
	if kind == "file" {
		return identifyFile(s, r, fsh, w, h, z, do, dirs)
	}
	// else: "directory"...
	dirs = append(dirs, fsh.Get("name").String())
	entries := fsh.Call("values")
	for {
		next := entries.Call("next")
		entry, err := await(next)
		if err != nil {
			return err
		}
		if entry.Get("done").Bool() {
			return nil
		}
		fsh := entry.Get("value")
		err = identifyFiles(s, r, fsh, w, h, z, do, dirs)
		if err != nil {
			return err
		}
	}
}

// identify(FileSystemHandle, ...OPTS)
// OPTS: json (default), csv, yaml, droid,
//
//	     md5, sha1, sha256, sha512, crc,
//		 z
func sfWrapper() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) < 1 {
			panic("SF WASM error: provide a FileSystemHandle as first argument")
		}

		fmt.Println(sfcontent[:10], len(sfcontent))
		sf := static.Newx(sfcontent)

		ls()
		//log.Println(sfcontent)

		promiseHandler := js.FuncOf(func(v js.Value, x []js.Value) interface{} {
			resolve := x[0]
			reject := x[1]
			go func() {
				// return a tuple of options that we can
				// provide siegfried with...
				o, ht, z := opts(args[1:])
				h := checksum.MakeHash(ht)
				out := &bytes.Buffer{}
				// writer I think is a result writer that means
				// we can write the results from SF somewhere...
				var w writer.Writer
				r := newReader()
				switch o {
				case yamlOut:
					w = writer.YAML(out)
				case csvOut:
					w = writer.CSV(out)
				case droidOut:
					w = writer.Droid(out)
				default:
					w = writer.JSON(out)
				}

				w.Head(config.SignatureBase(), time.Now(), sf.C, config.Version(), sf.Identifiers(), sf.Fields(), ht.String())

				//var err error
				//fmt.Println(z, h, r)

				// This is where the magic happens and we can
				// begin to identify the files...
				//
				// sf = a siegfried
				// r = a reader
				// args[0] = file system handler...
				// w = a writer...
				// h = a hash...
				// z = read archives...
				// o = result serialization...
				// dirs == nil...
				fsh := args[0]
				err := identifyFiles(sf, r, fsh, w, h, z, o == droidOut, nil)
				w.Tail()
				if err != nil {
					reject.Invoke(err.Error())
				} else {

					x := out
					y, err := json.MarshalIndent(x, "", "    ")
					if err != nil {
						panic(err)
					}
					fmt.Println(y)
					resolve.Invoke(out.String())
				}
			}()
			return nil
		})
		// Create and return the Promise object
		promise := js.Global().Get("Promise")
		return promise.New(promiseHandler)
	})
}

func ls() {
	files, err := filepath.Glob("*")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("files:", files) // contains a list of all files in the current directory
}

func something(extension []byte) error {
	opts := []config.Option{}
	id, err := pronom.Newx(container, droid, extension, opts...)
	if err != nil {
		fmt.Println(err)
		return err
	}
	s := siegfried.New()
	err = s.Add(id)
	if err != nil {
		fmt.Println(err)
		return err
	}
	var b bytes.Buffer
	foo := bufio.NewWriter(&b)
	err = s.Savex(foo)
	if err != nil {
		fmt.Println("error saving sf: ", err)
		return err
	}
	foo.Flush()
	x := b.Bytes()
	fmt.Println("have we saved? ", len(x))
	//fmt.Println(x[:10])
	sfcontent = x
	return nil
}

func royWrapper() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		promiseHandler := js.FuncOf(func(v js.Value, x []js.Value) interface{} {

			// x is a slice containing two functions
			// resolve and reject... they are called later.
			// v doesn't seem to be defined when it is called...
			resolve := x[0]
			reject := x[1]
			fmt.Println("vvv", v)
			fmt.Println("xxx", x)
			go func() {
				out := &bytes.Buffer{}
				//var w writer.Writer
				//data := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
				//out = append(*out, data...)
				out.Write([]byte("Successfully loaded new signature file")) // written to the resolve function...
				var err error
				//err = fmt.Errorf("ffffffffffffffff")

				fsh := args[0]
				kind := fsh.Get("kind").String()

				if kind != "file" {
					// handle error...
				}

				x := []string{}
				name := filepath.Join(append(x, fsh.Get("name").String())...)

				fmt.Println(name)

				//r.Get()

				r := newReader()
				promise := fsh.Call("getFile")
				val, err := await(promise)
				if err != nil {
					fmt.Println("fffffffffffff: ", err)
					// handle error
				}

				s := []byte{}
				r.reset(val)
				fmt.Println("xml size: ", len(r.buf))
				//i, err := r.Read(s)

				//if err != nil {
				// handle error...
				//}
				//log.Println("read bytes: ", i, len(s))

				//s := []byte{r.ReadAll()}
				//fmt.Println(r)

				s, _ = r.Slice(0, int(r.sz))
				fmt.Println("data???? ", s[:10])
				something(s)

				ls()

				if err != nil {
					reject.Invoke(err.Error())
				} else {
					resolve.Invoke(out.String())
				}
				fmt.Println("xxx works!") // written to console...

				//sfcontent = sf2
				fmt.Println(sfcontent[:10], len(sfcontent))

				//[]byte{0xfd, 0x5f, 0xff}

			}()
			return nil
		})
		// Create and return the Promise object
		promise := js.Global().Get("Promise")
		return promise.New(promiseHandler)
	})
}

func main() {
	js.Global().Set("identify", sfWrapper())
	js.Global().Set("sigload", royWrapper())
	<-make(chan bool)
}

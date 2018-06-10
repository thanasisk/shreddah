/*
a shred(1)-like utility, implemented in Golang

*/
package main

import "log"
import "os"
import "syscall"
import "unsafe"
import "io"
import "bytes"

// before the crypto purists start a pitchfork mob - we need some bytes
// for non-crypto purposes
import "math/rand"
import flags "github.com/jessevdk/go-flags"

var opts struct {
	//to be implemented in another version
	//Verbose bool `short:"v" long:"verbose" description:"verbose mode"`
	Unlink bool `short:"u" long:"unlink" description:"unlink file"`
	Passes uint `short:"p" long:"pass" default:"3" description:"number of passes"`
	Force  bool `short:"f" long:"force" description:"force chmod"`
	Zero   bool `short:"z" long:"zero" description:"do a last pass with zeroes"`
}

const AlignSize = 4096
const BlockSize = 4096

// shamelessly stolen
func alignment(block []byte, AlignSize int) int {
	return int(uintptr(unsafe.Pointer(&block[0])) & uintptr(AlignSize-1))
}

// shamelessly stolen as well
func AlignedBlock(BlockSize int) []byte {
	block := make([]byte, BlockSize+AlignSize)
	if AlignSize == 0 {
		return block
	}
	a := alignment(block, AlignSize)
	offset := 0
	if a != 0 {
		offset = AlignSize - a
	}
	block = block[offset : offset+BlockSize]
	// Can't check alignment of a zero sized block
	if BlockSize != 0 {
		a = alignment(block, AlignSize)
		if a != 0 {
			log.Fatal("Failed to align block")
		}
	}
	return block
}

// good coders create, great coders steal - was it Picasso or Dali?
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func overwrite(f *os.File, sz_file int64, pattern []byte) (int, error) {
	out := io.Writer(f)
	block := AlignedBlock(BlockSize)
	chunks := int64(0)
	if sz_file%int64(len(block)) > 0 {
		chunks++
	}
	for i := int64(0); i < chunks; i++ {
		if _, err := out.Write(pattern); err != nil {
			return -1, err
		}
	}
	f.Sync()
	return 0, nil
}

func shred(fname string) (int, error) {
	// sanity check if it makes sense to continue
	fi, err := os.Stat(fname)
	if err != nil {
		return -1, err
	}
	mode := fi.Mode()
	// maybe there is a shorter way of writing this ...
	if mode&os.ModeNamedPipe != 0 {
		log.Fatal("named pipe shredding NOT supported")
	}
	if mode&os.ModeSocket != 0 {
		log.Fatal("socket shredding NOT supported")
	}
	if mode&os.ModeTemporary != 0 {
		log.Fatal("Plan9 is NOT supported")
	}
	if mode&os.ModeDevice != 0 && mode&os.ModeCharDevice != 0 {
		log.Fatal("Character device shredding NOT supported")
	}
	// the following is Linux specific
	f, err := os.OpenFile(fname, syscall.O_DIRECT|os.O_WRONLY, 0660)
	if err != nil {
		if os.IsPermission(err) {
			if opts.Force == true {
				if err := os.Chmod(fname, fi.Mode()|syscall.S_IWUSR); err != nil {
					return -1, err

				}
			}
		} else {
			return -1, err
		}
	}
	// pray!
	defer f.Close()
	block := AlignedBlock(BlockSize)
	pattern := make([]byte, len(block))
	for i := uint(0); i < opts.Passes; i++ {
		rand.Read(pattern)
		if _, err := overwrite(f, fi.Size(), pattern); err != nil {
			return -1, err
		}
		f.Seek(0, 0)
	}
	if opts.Zero {
		f.Seek(0, 0)
		buf := bytes.Repeat([]byte{0x00}, len(block))
		if _, err := overwrite(f, fi.Size(), buf); err != nil {
			return -1, err
		}
	}
	return 0, nil
}

func unlink(fname string) error {
	// how atomic rename is?
	for i := len(fname); i > 1; i-- {
		newName := randomString(i)
		os.Rename(fname, newName)
		fname = newName
	}
	if err := os.Remove(fname); err != nil {
		return err
	}
	return nil
}

func main() {
	args, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		log.Fatal(err)
	}
	if len(args) < 2 {
		log.Fatal("we need at least 1 filename")
	}
	fnames := args[1:]
	for _, fname := range fnames {
		if _, err := shred(fname); err != nil {
			log.Fatal(err)
		}
		if opts.Unlink == true {
			if err := unlink(fname); err != nil {
				log.Fatal(err)
			}
		}
	}
}

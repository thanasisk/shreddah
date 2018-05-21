# Shreddah

## a (partial) GNU shred(1) implementation in Golang

## Usage
```shreddah -h``` will show you the options
```Application Options:
  -u, --unlink  unlink file after shredding
  -p, --pass=   number of passes (default: 3)
  -f, --force   force chmod if file is not writable
  -h, --help    Show this help message
```


## Platforms
Currently ONLY tested on Linux, with FreeBSD tests to follow. It will *NOT* run on your Windows/OSX
## Caveat Emptor
as stated in the man page of shred(1), there are a multitude of scenarios where
shreddah (or shred(1) for that matter) will *NOT* yield the expected results.
Shreddah was coded mostly as coding training, in a quiet weekend, thus don't base
your anti-forensic strategy on it.
## Contributions
Contributions are welcome - make sure you follow CoC
## Bugs
probably a few, drop me a line or open an issue or (even better!) fix and submit PR
## License
*GNU General Public License v2.0*

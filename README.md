Deps install:

    brew install lua
    brew link --overwrite lua
    cd /usr/local/lib
    ln -s liblua.5.1.5.dylib liblua.dylib


Go&Code install:

    brew install go
    export GOPATH=~
    go get github.com/nordicdyno/luawshop2014-gopher-on-lua
    ~/bin/luawshop2014-gopher-on-lua
    

#!/bin/bash

unit_tests(){
    local ret
    ret=0
    for d in $(go list ./... | grep -v vendor); do 
        COV=$(go test -race  -coverprofile=profile.out -covermode=atomic $d | sed 's/.*coverage//g' | sed  's/[^0-9.]*//g') 
        if [ -f profile.out ]; then 
            cat profile.out >> coverage.txt;
            rm profile.out;
        fi 
        if [[ $COV < 75.0 ]]; then  
            echo expecting test coverage greater than 75 %, got insufficient $COV % for package $d; 
            if test $ret -eq 0; then
                ret=1
            fi
        fi
    done
    echo $ret
    return $ret
}

set -e
unit_tests
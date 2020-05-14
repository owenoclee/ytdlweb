// +build ignore

package main

import (
    "log"
    "net/http"

    "github.com/shurcooL/vfsgen"
)

func main() {
    err := vfsgen.Generate(
        http.Dir("./public"),
        vfsgen.Options{
            Filename:     "./public_assets.go",
            PackageName:  "main",
            VariableName: "publicAssets",
        },
    )
    if err != nil {
        log.Fatalln(err)
    }
}


// Copyright (c) 2014, fromkeith
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without modification,
// are permitted provided that the following conditions are met:
//
// * Redistributions of source code must retain the above copyright notice, this
//   list of conditions and the following disclaimer.
//
// * Redistributions in binary form must reproduce the above copyright notice, this
//   list of conditions and the following disclaimer in the documentation and/or
//   other materials provided with the distribution.
//
// * Neither the name of the fromkeith nor the names of its
//   contributors may be used to endorse or promote products derived from
//   this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
// ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
// (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
// LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
// ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
// SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
///

package gorest


import (
    "reflect"
    "strings"
    "html/template"
    "bytes"
    "log"
    "go/token"
    "io/ioutil"
    "path/filepath"
    "go/doc"
    "go/ast"
    "go/parser"
    "go/build"
)


func docImporter(imports map[string]*ast.Object, path string) (*ast.Object, error) {
    pkg := imports[path]
    if pkg == nil {
        name := path[strings.LastIndex(path, "/")+1:]
        pkg = ast.NewObj(ast.Pkg, name)
        pkg.Data = ast.NewScope(nil) // required by ast.NewPackage for dot-import
        imports[path] = pkg
    }
    return pkg, nil
}


func extractComments(packageLocation string) map[string]string {
    bpkg, err := build.Default.Import("path/to/test/package", ".", 0)
    if err != nil {
        log.Fatal(err)
    }
    fset := token.NewFileSet()
    files := make(map[string]*ast.File)
    for _, fname := range bpkg.GoFiles {
        p, err := ioutil.ReadFile(filepath.Join(bpkg.SrcRoot, bpkg.ImportPath, fname))
        if err != nil {
            log.Fatal(err)
        }
        file, err := parser.ParseFile(fset, fname, p, parser.ParseComments)
        if err != nil {
            log.Fatal(err)
        }
        files[fname] = file
    }

    log.Printf("GoFiles: %v\n", files["orders.putAlbumOrder.go"].Comments[0].End())
    result := make(map[string]string)

    pkg, _ := ast.NewPackage(fset, files, docImporter, nil)

    dpkg := doc.New(pkg, bpkg.ImportPath, 0)
    for i := range dpkg.Types {
        for j := range dpkg.Types[i].Methods {
            result[strings.ToLower(dpkg.Types[i].Name + "." + dpkg.Types[i].Methods[j].Name)] = dpkg.Types[i].Methods[j].Doc
        }
    }
    return result
}



/*
    Creates documentation about the specified service.
    h - the service to document
    outputTemplate - the html template to use. See doc.template.html
    returns the string template
*/
func DocumentService(h interface{}, root string, outputTemplate *template.Template, sourceFileLocation string) string {
    if _, ok := h.(GoRestService); !ok {
        panic(ERROR_INVALID_INTERFACE)
    }
    t := reflect.TypeOf(h)

    if t.Kind() == reflect.Ptr {
        t = t.Elem()
    } else {
        panic(ERROR_INVALID_INTERFACE)
    }
    docManager := new(manager)
    docManager.serviceTypes = make(map[string]serviceMetaData, 0)
    docManager.endpoints = make(map[string]endPointStruct, 0)
    docManager.logger = defaultLogger{}

    if t.Kind() == reflect.Struct {
        if field, found := t.FieldByName("RestService"); found {
            temp := strings.Join(strings.Fields(string(field.Tag)), " ")
            meta := prepServiceMetaData(docManager, root, reflect.StructTag(temp), h, t.Name())
            tFullName := docManager.addType(t.PkgPath()+"/"+t.Name(), meta)
            for i := 0; i < t.NumField(); i++ {
                f := t.Field(i)
                mapFieldsToMethods(docManager, t, f, tFullName, meta)
            }

            if len(docManager.serviceTypes) != 1 {
                panic("Only expected 1 service")
            }
            comments := extractComments(sourceFileLocation)


            service := docManager.serviceTypes[tFullName]
            var doc docOutput
            doc.ServiceName = tFullName
            doc.Endpoints = make([]docEndpoint, 0, len(docManager.endpoints))
            for _, v := range docManager.endpoints {
                // reflect the service so we can get the types in the endpoint
                servVal := reflect.ValueOf(h)
                targetMethod := servVal.Type().Method(v.methodNumberInParent)
                var postDataDescription *docStruct
                if v.postdataType != "" {
                    postDataDescription = docDescribeStruct(targetMethod.Type.In(1))
                }
                var outputDescription *docStruct
                if v.outputType != "" {
                    outputDescription = docDescribeStruct(targetMethod.Type.Out(0))
                }

                consumeMime := service.consumesMime
                if v.overrideConsumesMime != "" {
                    consumeMime = v.overrideConsumesMime
                }
                prodMime := service.producesMime
                if v.overrideProducesMime != "" {
                    prodMime = v.overrideProducesMime
                }

                methodDoc, _ := comments[strings.ToLower(t.Name() + "." + v.name)]

                doc.Endpoints = append(doc.Endpoints, docEndpoint{
                    Name: v.name,
                    RequestMethod: v.requestMethod,
                    Signature: v.signiture,
                    Root: v.root,
                    Params: v.params,
                    QueryParams: v.queryParams,
                    SignitureLen: v.signitureLen,
                    ParamLen: v.paramLen,
                    InputMime: v.inputMime,
                    OutputType: v.outputType,
                    OutputTypeIsArray: v.outputTypeIsArray,
                    OutputTypeIsMap: v.outputTypeIsMap,
                    PostdataType: v.postdataType,
                    PostdataTypeIsArray: v.postdataTypeIsArray,
                    PostdataTypeIsMap: v.postdataTypeIsMap,
                    IsVariableLength: v.isVariableLength,
                    ParentTypeName: v.parentTypeName,
                    MethodNumberInParent: v.methodNumberInParent,
                    Role: v.role,
                    OverrideProducesMime: prodMime,
                    OverrideConsumesMime: consumeMime,
                    AllowGzip: v.allowGzip,
                    PostData: postDataDescription,
                    Output: outputDescription,
                    Doc: methodDoc,
                })
            }

            buf := bytes.Buffer{}
            outputTemplate.ExecuteTemplate(&buf, "gorest.service", doc)
            return buf.String()
        }
        return ""
    }
    panic(ERROR_INVALID_INTERFACE)
    return ""
}

func docDescribeStruct(dataType reflect.Type) *docStruct {
    var desc docStruct
    if dataType.Kind() == reflect.Slice {
        desc.IsArray = true
        dataType = dataType.Elem()
    }
    inst := reflect.New(dataType).Elem().Type()
    desc.Fields = make([]docField, inst.NumField())
    for i := 0; i < inst.NumField(); i++ {
        desc.Fields[i] = docField{
            Name: inst.Field(i).Name,
            Type: inst.Field(i).Type.String(),
        }
    }
    return &desc
}

type docField struct {
    Name            string
    Type            string
}

type docStruct struct {
    Fields              []docField
    IsArray             bool
}

type docOutput struct {
    ServiceName         string
    Endpoints           []docEndpoint
}

type docEndpoint struct {
    Name                 string
    RequestMethod        string
    Signature            string
    MuxRoot              string
    Root                 string
    NonParamPathPart     map[int]string
    Params               []param //path parameter name and position
    QueryParams          []param
    SignitureLen         int
    ParamLen             int
    InputMime            string
    OutputType           string
    OutputTypeIsArray    bool
    OutputTypeIsMap      bool
    PostdataType         string
    PostdataTypeIsArray  bool
    PostdataTypeIsMap    bool
    IsVariableLength     bool
    ParentTypeName       string
    MethodNumberInParent int
    Role                 string
    OverrideProducesMime string // overrides the produces mime type
    OverrideConsumesMime string // overrides the produces mime type
    AllowGzip            int // 0 false, 1 true, 2 unitialized
    PostData                *docStruct
    Output                  *docStruct
    Doc                     string
}
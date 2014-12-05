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
    "log"
    "go/token"
    "io/ioutil"
    "path/filepath"
    "go/doc"
    "go/ast"
    "go/parser"
    "go/build"
    "runtime"
    "os"
    "sort"
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

// visits a struct to extract the comments on fields
type vis struct {
    feildComments       map[string]string
}
// visits a field to extract its comments
type fieldVis struct {
    fieldName           string
    v                   *vis
}
func (v fieldVis) Visit(node ast.Node) (w ast.Visitor) {
    if c, ok := node.(*ast.CommentGroup); ok {
        v.v.feildComments[v.fieldName] = c.Text()
    }
    // TODO: look at the tags. as json may rename fields. ditto for xml.
    return v
}

func (v *vis) Visit(node ast.Node) (w ast.Visitor) {
    switch node := node.(type) {
        case *ast.Field:
            if len(node.Names) == 0 {
                return v
            }
            return fieldVis{fieldName: node.Names[0].Name, v: v}
        case *ast.GenDecl, *ast.TypeSpec, *ast.StructType, *ast.FieldList, *ast.Ident:
            return v
    }
    return nil
}


func extractComments(packageImportPath string) map[string]string {
    bpkg, err := build.Default.Import(packageImportPath, ".", 0)
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

    result := make(map[string]string)

    pkg, _ := ast.NewPackage(fset, files, docImporter, nil)

    dpkg := doc.New(pkg, bpkg.ImportPath, 0)
    for i := range dpkg.Types {
        result[dpkg.Types[i].Name] = dpkg.Types[i].Doc
        for j := range dpkg.Types[i].Methods {
            result[dpkg.Types[i].Name + "." + dpkg.Types[i].Methods[j].Name] = dpkg.Types[i].Methods[j].Doc
        }
        // visit the struct extracting any comments on fields
        v := new(vis)
        v.feildComments = make(map[string]string)
        ast.Walk(v,  dpkg.Types[i].Decl)
        for fieldName, feildComment := range v.feildComments {
            if feildComment == "" {
                continue
            }
            result[dpkg.Types[i].Name + "." + fieldName] = feildComment
        }
    }
    for i := range dpkg.Funcs {
        result[dpkg.Funcs[i].Name] = dpkg.Funcs[i].Doc
    }
    return result
}


/*
    Creates HTML documents of all the registered services and authorizers.
    Register your service as you normally would.
        gorest.RegisterService(new(MyService))

    Load up the template files into a string
        buf := bytes.Buffer{}
        buf.WriteString(readInFile("src/github.com/fromkeith/gorest/server.doc.template.html"))
        buf.WriteString("\n")
        buf.WriteString(readInFile("src/github.com/fromkeith/gorest/auth.doc.template.html"))
        buf.WriteString("\n")
        buf.WriteString(readInFile("src/github.com/fromkeith/gorest/structs.doc.template.html"))
        buf.WriteString("\n")
        buf.WriteString(readInFile("src/github.com/fromkeith/gorest/index.doc.template.html"))

    Then call this function.
        gorest.DocumentServices(buf.String(), "outputFolder")

    Outfolder will be populated with nice documentation on your service!
*/
func DocumentServices(templateSource string, outputFolder string) {
    packageComments := make(map[string]map[string]string)
    services := make(map[string]docOutput)
    auths := make(map[string]docAuth)

    manager := _manager()

    for fullName, servMeta := range manager.serviceTypes {
        t := reflect.TypeOf(servMeta.template).Elem()
        if _, ok := packageComments[t.PkgPath()]; !ok {
            packageComments[t.PkgPath()] = extractComments(t.PkgPath())
        }
        services[fullName] = docOutput{
            Service: docService{
                Name : t.Name(),
                Doc : packageComments[t.PkgPath()][t.Name()],
                Realm : servMeta.realm,
                Root: servMeta.root,
                meta: servMeta,
            },
            Endpoints: make([]docEndpoint, 0, 10),
        }
    }
    for _, v := range manager.endpoints {
        var s docOutput
        var ok bool
        if s, ok = services[v.parentTypeName]; !ok {
            panic("CouldNotFindServiceForEndPoint")
        }
        ep := documentEndpoint(v, s.Service, manager, packageComments)
        s.Endpoints = append(s.Endpoints, ep)
        services[v.parentTypeName] = s
    }

    for name, auther := range authorizers {
        funcName := runtime.FuncForPC(reflect.ValueOf(auther).Pointer()).Name()
        if funcName == "" {
            continue
        }
        lastSlash := strings.LastIndex(funcName, "/")
        firstDot := strings.Index(funcName[lastSlash:], ".")
        pkgPath := funcName[:lastSlash + firstDot]
        actualFuncName := funcName[lastSlash + firstDot + 1:]
        if _, ok := packageComments[pkgPath]; !ok {
            packageComments[pkgPath] = extractComments(pkgPath)
        }

        auths[name] = docAuth{
            Realm: name,
            Doc: packageComments[pkgPath][actualFuncName],
        }
    }
    structList := make([]docStruct, 0, len(manager.endpoints))
    for i := range services {
        for j := range services[i].Endpoints {
            if services[i].Endpoints[j].PostData != nil {
                structList = append(structList, *services[i].Endpoints[j].PostData)
            }
            if services[i].Endpoints[j].Output != nil {
                structList = append(structList, *services[i].Endpoints[j].Output)
            }
        }
    }
    outputResults(templateSource, outputFolder, services, auths, structList)

}

type sortedStructs []docStruct

func (a sortedStructs) Len() int           { return len(a) }
func (a sortedStructs) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a sortedStructs) Less(i, j int) bool { return a[i].Name < a[j].Name }

type sortedAuths []docAuth

func (a sortedAuths) Len() int           { return len(a) }
func (a sortedAuths) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a sortedAuths) Less(i, j int) bool { return a[i].Realm < a[j].Realm }

func outputResults(templateSource string, outputFolder string, services map[string]docOutput, auths map[string]docAuth, structList []docStruct) {
    sl := sortedStructs(structList)
    sort.Sort(sl)

    authList := make([]docAuth, 0, len(auths))
    for _, v := range auths {
        authList = append(authList, v)
    }

    sa := sortedAuths(authList)
    sort.Sort(sa)


    tmpl, err := template.New("gorest").Parse(templateSource)
    if err != nil {
        panic(err)
    }

    for _, v := range services {
        outFile, err := os.Create(filepath.Join(outputFolder, "service." + v.Service.Name + ".html"))
        if err != nil {
            panic(err)
        }
        defer outFile.Close()
        err = tmpl.ExecuteTemplate(outFile, "gorest.service", v)
        if err != nil {
            panic(err)
        }
    }
    authOutFile, err := os.Create(filepath.Join(outputFolder, "auths.html"))
    if err != nil {
        panic(err)
    }
    defer authOutFile.Close()
    err = tmpl.ExecuteTemplate(authOutFile, "gorest.auths", docAuthContainer{authList})
    if err != nil {
        panic(err)
    }
    dataTypeOutFile, err := os.Create(filepath.Join(outputFolder, "datatypes.html"))
    if err != nil {
        panic(err)
    }
    defer dataTypeOutFile.Close()
    err = tmpl.ExecuteTemplate(dataTypeOutFile, "gorest.structs", docDataTypes{structList})
    if err != nil {
        panic(err)
    }
    serviceNames := make([]string, 0, len(services))
    for _, v := range services {
        serviceNames = append(serviceNames, v.Service.Name)
    }

    sort.Strings(serviceNames)
    indexOutFile, err := os.Create(filepath.Join(outputFolder, "index.html"))
    if err != nil {
        panic(err)
    }
    defer indexOutFile.Close()
    err = tmpl.ExecuteTemplate(indexOutFile, "gorest.index", docIndex{serviceNames})
    if err != nil {
        panic(err)
    }
}

func documentEndpoint(v endPointStruct, service docService, manager *manager, packageComments map[string]map[string]string) docEndpoint {
    servVal := reflect.ValueOf(service.meta.template)
    t := reflect.TypeOf(service.meta.template).Elem()
    targetMethod := servVal.Type().Method(v.methodNumberInParent)
    var postDataDescription *docStruct
    if v.postdataType != "" {
        postDataDescription = docDescribeStruct(targetMethod.Type.In(1), packageComments)
    }
    var outputDescription *docStruct
    if v.outputType != "" {
        outputDescription = docDescribeStruct(targetMethod.Type.Out(0), packageComments)
    }

    consumeMime := service.meta.consumesMime
    if v.overrideConsumesMime != "" {
        consumeMime = v.overrideConsumesMime
    }
    prodMime := service.meta.producesMime
    if v.overrideProducesMime != "" {
        prodMime = v.overrideProducesMime
    }

    name := strings.ToUpper(v.name[0:1]) + v.name[1:]
    methodDoc := packageComments[t.PkgPath()][t.Name() + "." + name]

    return docEndpoint{
        Name: v.name,
        RequestMethod: v.requestMethod,
        MethodDefaultReturn: getDefaultResponseCode(v.requestMethod),
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
        ProducesMime: prodMime,
        ConsumesMime: consumeMime,
        AllowGzip: v.allowGzip,
        PostData: postDataDescription,
        Output: outputDescription,
        Doc: methodDoc,
    }
}

func docDescribeStruct(dataType reflect.Type, packageComments map[string]map[string]string) *docStruct {
    var desc docStruct
    if dataType.Kind() == reflect.Slice {
        desc.IsArray = true
        dataType = dataType.Elem()
    }
    inst := reflect.New(dataType).Elem().Type()
    if _, ok := packageComments[inst.PkgPath()]; !ok {
        packageComments[inst.PkgPath()] = extractComments(inst.PkgPath())
    }
    desc.Name = dataType.Name()
    desc.Doc = packageComments[inst.PkgPath()][desc.Name]
    desc.Fields = make([]docField, inst.NumField())
    for i := 0; i < inst.NumField(); i++ {
        desc.Fields[i] = docField{
            Name: inst.Field(i).Name,
            Type: inst.Field(i).Type.String(),
            Doc: packageComments[inst.PkgPath()][desc.Name + "." + inst.Field(i).Name],
        }
    }
    return &desc
}

type docField struct {
    Name            string
    Type            string
    Doc             string
}

type docDataTypes struct {
    Structs         []docStruct
}

type docStruct struct {
    Fields              []docField
    IsArray             bool
    Name                string
    Doc                 string
}

type docService struct {
    Name            string
    Realm           string
    Doc             string
    Root            string

    meta            serviceMetaData
}

type docAuth struct {
    Realm           string
    Doc             string
}
type docAuthContainer struct {
    Auths               []docAuth
}

type docOutput struct {
    Service             docService
    Endpoints           []docEndpoint
}

type docIndex struct {
    Services            []string
}

type docEndpoint struct {
    Name                    string
    RequestMethod           string
    MethodDefaultReturn     int
    Signature               string
    MuxRoot                 string
    Root                    string
    NonParamPathPart        map[int]string
    Params                  []param //path parameter name and position
    QueryParams             []param
    SignitureLen            int
    ParamLen                int
    InputMime               string
    OutputType              string
    OutputTypeIsArray       bool
    OutputTypeIsMap         bool
    PostdataType            string
    PostdataTypeIsArray     bool
    PostdataTypeIsMap       bool
    IsVariableLength        bool
    ParentTypeName          string
    MethodNumberInParent    int
    Role                    string
    ProducesMime            string // overrides the produces mime type
    ConsumesMime            string // overrides the produces mime type
    AllowGzip               int // 0 false, 1 true, 2 unitialized
    PostData                *docStruct
    Output                  *docStruct
    Doc                     string
}
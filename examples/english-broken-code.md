# 4.2 Protobuf

Protobuf is short for Protocol Buffers, a data description language developed by Google and opened to the public in 2008. Protobuf's positioning when it was just open source is similar to XML, JSON and other data description languages. It generates code through the accompanying tools and implements the function of serializing structured data. But we are more concerned about Protobuf as the description language of the interface specification, which can be used as the basic tool for designing a secure cross-language PRC interface.

## 4.2.1 Getting Started with Protobuf

For readers who have not used Protobuf, it is recommended to understand the basic usage from the official website.Here we try to combine Protobuf and RPC, and finally guarantee the interface specification and security of RPC through Protobuf. The most basic unit of data in Protobuf is the message, which is similar to the structure of the Go language.Members of the message or other underlying data types can be nested in the message.

First create a hello.proto file that wraps the string type used in the HelloService service:

```protobuf
syntax = "proto3" ; package main; message String { string value = 1 ; } 
```

The syntax statement at the beginning indicates the syntax of using proto3. The third version of Protobuf simplifies the language, and all members are initialized with zero values like Go (no custom defaults are supported), so message members no longer need to support the required attribute. Then the package directive indicates that the main package is currently (this can be consistent with the Go package name, simplifying the example code), of course, the user can also customize the corresponding package path and name for different languages. Finally, the message keyword defines a new String type, which corresponds to a String structure in the final generated Go language code. There is only one value member of the string type in the String type, and the member is encoded with a 1 number instead of the name.

In a data description language such as XML or JSON, the corresponding data is generally bound by the name of the member. However, Protobuf encoding binds the corresponding data by the unique number of the member, so the data size of Protobuf encoding will be smaller, but it is also very inconvenient for humans to consult. We are not currently concerned with Protobuf's encoding technology. The resulting Go structure can be freely encoded in JSON or gob, so you can temporarily ignore the member encoding part of Protobuf.

The Protobuf core toolset is developed in the C++ language and does not support the Go language in the official protoc compiler. To generate the corresponding Go code based on the above hello.proto file, you need to install the appropriate plugin. The first is to install the official protoc tool, which can be downloaded from [https://github.com/google/protobuf/releases](https://translate.googleusercontent.com/translate_c?depth=1&hl=en&rurl=translate.google.com&sl=zh-CN&sp=nmt4&tl=en&u=https://github.com/google/protobuf/releases&xid=17259,15700023,15700124,15700149,15700186,15700190,15700201&usg=ALkJrhgn8ntqXbaGhxRNclwd8y4XkbGxMg) . Then install the code generation plugin for Go, which can be installed via the `go get github.com/golang/protobuf/protoc-gen-go` command.

Then generate the corresponding Go code with the following command:

```
 $ protoc --go_out=. hello.proto 
```

The `go_out` parameter tells the protoc compiler to load the corresponding protoc-gen-go tool, then generate the code through the tool and generate the code into the current directory. Finally, there is a list of a series of protobuf files to process.

Only a hello.pb.go file is generated here, where the String structure is as follows:

```go
 type String struct { Value string `protobuf:"bytes,1,opt,name=value" json:"value,omitempty"` } func (m *String) Reset() { *m = String{} } func (m *String) String() string { return proto.CompactTextString(m) } func (*String) ProtoMessage() {} func (*String) Descriptor() ([] byte , [] int ) { return fileDescriptor_hello_069698f99dd8f029, [] int { 0 } } func (m *String) GetValue() string { if m != nil { return m.Value } return "" } 
```

The generated structure will also contain some members prefixed with `XXX_` , which we have hidden. At the same time, the String type also automatically generates a set of methods, of which the ProtoMessage method indicates that this is a method that implements the proto.Message interface. In addition, Protobuf generates a Get method for each member. The Get method can handle not only the null pointer type, but also the Protobuf version 2 method (the second version of the custom default feature depends on this method).

Based on the new String type, we can reimplement the HelloService service:

```go
 type HelloService struct {} func (p *HelloService) Hello(request *String, reply *String) error { reply.Value = "hello:" + request.GetValue() return nil } 
```

The input parameters and output parameters of the Hello method are all represented by the String type defined by Protobuf. Because the new input parameter is a structure type, the pointer type is used as the input parameter, and the internal code of the function is also adjusted accordingly.

So far, we have initially realized the combination of Protobuf and RPC. When starting the RPC service, we can still choose the default gob or manually specify the json code, or even re-implement a plugin based on the protobuf code.Although I have done so much work, it seems that I have not seen any gains!

Looking back at the more secure RPC interface part of Chapter 1, we spent a great deal of effort to add security to RPC services. The resulting code for the more secure RPC interface itself is very cumbersome to use manual maintenance, while all security-related code is only available for the Go language environment! Since the input and output parameters defined by Protobuf are used, can the RPC service interface be defined by Protobuf? Its practical Protobuf defines the language-independent RPC service interface is its real value!

Update the hello.proto file below to define the HelloService service via Protobuf:

```protobuf
 service HelloService { rpc Hello (String) returns (String) ; } 
```

But the regenerated Go code hasn't changed. This is because there are millions of RPC implementations in the world, and the protoc compiler does not know how to generate code for the HelloService service.

However, a plugin called grpc has been integrated inside protoc-gen-go to generate code for grpc:

```
 $ protoc --go_out=plugins=grpc:. hello.proto 
```

In the generated code, there are some new types like HelloServiceServer and HelloServiceClient. These types are for grpc and do not meet our RPC requirements.

However, the grpc plugin provides us with an improved idea. Below we will explore how to generate secure code for our RPC.

## 4.2.2 Custom Code Generation Plugin

Protobuf's protoc compiler implements support for different languages through a plugin mechanism. For example, if the protoc command appears in the format of `--xxx_out` format, then protoc will first query whether there is a built-in xxx plugin. If there is no built-in xxx plugin, it will continue to query whether there is a protoc-gen-xxx named executable program in the current system. Generate code by querying the plugin. For the Protoc-gen-go plugin for Go, there is a layer of static plugin system. For example, protoc-gen-go has a built-in grpc plugin. Users can generate grpc related code by using the `--go_out=plugins=grpc` parameter. Otherwise, only relevant code will be generated for the message.

Referring to the code of the grpc plugin, you can find that the generator.RegisterPlugin function can be used to register the plugin. The plugin is a generator.Plugin interface:

```go
 // A Plugin provides functionality to add to the output during // Go code generation, such as to produce RPC stubs. type Plugin interface { // Name identifies the plugin. Name() string // Init is called once after data structures are built but before // code generation begins. Init(g *Generator) // Generate produces the code generated by the plugin for this file, // except for the imports, by calling the generator's methods P, In, // and Out. Generate(file *FileDescriptor) // GenerateImports produces the import declarations for this file. // It is called after Generate. GenerateImports(file *FileDescriptor) } 
```

The Name method returns the name of the plugin. This is the plugin system for the Protobuf implementation of the Go language. It has nothing to do with the name of the protoc plugin. Then the Init function initializes the plugin with the g parameter, which contains all the information about the Proto file. The final Generate and GenerateImports methods are used to generate the body code and the corresponding import package code.

So we can design a netrpcPlugin plugin to generate code for the standard library's RPC framework:

```go
 import ( "github.com/golang/protobuf/protoc-gen-go/generator" ) type netrpcPlugin struct { *generator.Generator } func (p *netrpcPlugin) Name() string { return "netrpc" } func (p *netrpcPlugin) Init(g *generator.Generator) { p.Generator = g } func (p *netrpcPlugin) GenerateImports(file *generator.FileDescriptor) { if len (file.Service) > 0 { p.genImportCode(file) } } func (p *netrpcPlugin) Generate(file *generator.FileDescriptor) { for _, svc := range file.Service { p.genServiceCode(svc) } } 
```

First the Name method returns the name of the plugin. The netrpcPlugin plugin has an anonymous `*generator.Generator` member built in, and then initialized with the parameter g when Init is initialized, so the plugin inherits all public methods from the g parameter object. The GenerateImports method calls the custom genImportCode function to generate the import code. The Generate method calls the custom genServiceCode method to generate the code for each service.

Currently, the custom genImportCode and genServiceCode methods simply output a simple comment:

```go
 func (p *netrpcPlugin) genImportCode(file *generator.FileDescriptor) { pP( "// TODO: import code" ) } func (p *netrpcPlugin) genServiceCode(svc *descriptor.ServiceDescriptorProto) { pP( "// TODO: service code, Name = " + svc.GetName()) } 
```

To use the plugin, you need to register the plugin with the generator.RegisterPlugin function, which can be done in the init function:

```go
 func init() { generator.RegisterPlugin( new (netrpcPlugin)) } 
```

Because the Go language package can only be imported statically, we can't add our newly written plugin to the installed protoc-gen-go. We will re-clone the main function corresponding to protoc-gen-go:

```go
 package main import ( "io/ioutil" "os" "github.com/golang/protobuf/proto" "github.com/golang/protobuf/protoc-gen-go/generator" ) func main() { g := generator.New() data, err := ioutil.ReadAll(os.Stdin) if err != nil { g.Error(err, "reading input" ) } if err := proto.Unmarshal(data, g.Request); err != nil { g.Error(err, "parsing input proto" ) } if len (g.Request.FileToGenerate) == 0 { g.Fail( "no files to generate" ) } g.CommandLineParameters(g.Request.GetParameter()) // Create a wrapped version of the Descriptors and EnumDescriptors that // point to the file that defines them. g.WrapTypes() g.SetPackageNames() g.BuildTypeNameMap() g.GenerateAllFiles() // Send back the results. data, err = proto.Marshal(g.Response) if err != nil { g.Error(err, "failed to marshal output proto" ) } _, err = os.Stdout.Write(data) if err != nil { g.Error(err, "failed to write output proto" ) } } 
```

In order to avoid interference with the protoc-gen-go plugin, we named our executable program protoc-gen-go-netrpc, which means that the nerpc plugin is included. Then recompile the hello.proto file with the following command:

```
 $ protoc --go-netrpc_out=plugins=netrpc:. hello.proto 
```

The `--go-netrpc_out` parameter tells the protoc compiler to load a plugin named protoc-gen-go-netrpc. The `plugins=netrpc` indicates that the internally unique netrpcPlugin plugin named netrpc is enabled. The added comment code will be included in the newly generated hello.pb.go file.

At this point, the hand-customized Protobuf code generation plugin is finally working.

## 4.2.3 Automatic generation of complete RPC code

In the previous example we have built a minimalized netrpcPlugin plugin and created a new plugin for protoc-gen-go-netrpc by cloning the main program of protoc-gen-go. Now continue to improve the netrpcPlugin plugin, the ultimate goal is to generate an RPC security interface.

The first is the code that generates the import package in the custom genImportCode method:

```go
 func (p *netrpcPlugin) genImportCode(file *generator.FileDescriptor) { pP( `import "net/rpc"` ) } 
```

Then generate the relevant code for each service in the custom genServiceCode method. Analysis can find that the most important thing for each service is the name of the service, and then each service has a set of methods. For the service definition method, the most important is the name of the method, as well as the names of the input parameters and output parameter types.

For this we define a ServiceSpec type that describes the meta-information of the service:

```go
 type ServiceSpec struct { ServiceName string MethodList []ServiceMethodSpec } type ServiceMethodSpec struct { MethodName string InputTypeName string OutputTypeName string } 
```

Then we create a new buildServiceSpec method to parse the ServiceSpec meta information for each service:

```go
 func (p *netrpcPlugin) buildServiceSpec(svc *descriptor.ServiceDescriptorProto) *ServiceSpec { spec := &ServiceSpec{ ServiceName: generator.CamelCase(svc.GetName()), } for _, m := range svc.Method { spec.MethodList = append (spec.MethodList, ServiceMethodSpec{ MethodName: generator.CamelCase(m.GetName()), InputTypeName: p.TypeName(p.ObjectNamed(m.GetInputType())), OutputTypeName: p.TypeName(p.ObjectNamed(m.GetOutputType())), }) } return spec } 
```

The input parameter is of type `*descriptor.ServiceDescriptorProto` , which fully describes all the information of a service. Then you can get the name of the service defined in the Protobuf file through `svc.GetName()` . After the name in the Protobuf file is changed to the name of the Go language, it needs to be converted by the `generator.CamelCase`function. Similarly, in the for loop we get the name of the method via `m.GetName()` and then change to the corresponding name in the Go language. More complicated is the resolution of the input and output parameter names: first need to obtain the type of the input parameter through `m.GetInputType()` , then get the class object information corresponding to the type through `p.ObjectNamed` , and finally get the name of the class object.

Then we can generate the code of the service based on the meta information of the service constructed by the buildServiceSpec method:

```go
 func (p *netrpcPlugin) genServiceCode(svc *descriptor.ServiceDescriptorProto) { spec := p.buildServiceSpec(svc) var buf bytes.Buffer t := template.Must(template.New( "" ).Parse(tmplService)) err := t.Execute(&buf, spec) if err != nil { log.Fatal(err) } pP(buf.String()) } 
```

For ease of maintenance, we generate service code based on a Go language template, where tmplService is the template for the service.

Before writing a template, let's look at what the final code we expect to generate looks like:

```go
 type HelloServiceInterface interface { Hello(in String, out *String) error } func RegisterHelloService(srv *rpc.Server, x HelloService) error { if err := srv.RegisterName( "HelloService" , x); err != nil { return err } return nil } type HelloServiceClient struct { *rpc.Client } var _ HelloServiceInterface = (*HelloServiceClient)( nil ) func DialHelloService(network, address string ) (*HelloServiceClient, error) { c, err := rpc.Dial(network, address) if err != nil { return nil , err } return &HelloServiceClient{Client: c}, nil } func (p *HelloServiceClient) Hello(in String, out *String) error { return p.Client.Call( "HelloService.Hello" , in, out) } 
```

The HelloService is the service name, and there are a series of method-related names.

The following template can be built with reference to the final code to be generated:

```go
 const tmplService = ` {{$root := .}} type {{.ServiceName}}Interface interface { {{- range $_, $m := .MethodList}} {{$m.MethodName}}(*{{$m.InputTypeName}}, *{{$m.OutputTypeName}}) error {{- end}} } func Register{{.ServiceName}}( srv *rpc.Server, x {{.ServiceName}}Interface, ) error { if err := srv.RegisterName("{{.ServiceName}}", x); err != nil { return err } return nil } type {{.ServiceName}}Client struct { *rpc.Client } var _ {{.ServiceName}}Interface = (*{{.ServiceName}}Client)(nil) func Dial{{.ServiceName}}(network, address string) ( *{{.ServiceName}}Client, error, ) { c, err := rpc.Dial(network, address) if err != nil { return nil, err } return &{{.ServiceName}}Client{Client: c}, nil } {{range $_, $m := .MethodList}} func (p *{{$root.ServiceName}}Client) {{$m.MethodName}}( in *{{$m.InputTypeName}}, out *{{$m.OutputTypeName}}, ) error { return p.Client.Call("{{$root.ServiceName}}.{{$m.MethodName}}", in, out) } {{end}} ` 
```

When Protobuf's plugin customization is complete, the code can be automatically generated each time the RPC service changes in the hello.proto file. You can also adjust or increase the content of the generated code by updating the template of the plugin. After mastering the custom Protobuf plugin technology, you will have this technology completely.
/*
A RESTful style web-services framework for the Go language.

Creating services in Go is straight forward, GoRest takes this a step further by adding a layer that
makes tedious tasks much more automated and avoids regular pitfalls.
This gives you the opportunity to focus more on the task at hand... minor low-level http handling.

Service Definition

A rest service is defined using tags, and embedding the gorest.RestService struct.

    type HelloService struct {
        // defines the root of this service, and its meta data.
        gorest.RestService `root:"/tutorial/"`
        // an end point that can be hit by a web request
        helloWorld      gorest.EndPoint `method:"GET" path:"/hello-world/" output:"string"`
    }

The service above defines the endpoint "/tutorial/hello-world/". It accepts the "GET" method.
It outputs a value of type string. The variable 'helloWorld' also specifies the name of the function
that gorest expects to exist. For the above service to run, you must implement the following definition:

    func (h HelloService) HelloWorld() string {
        return "Hello World"
    }

After defining the service, we next need to register it.

    func main() {
        // This tells gorest about the existence of the HelloService
        gorest.RegisterService(new(HelloService))
        // This allows gorest to handle the incoming http request on that root
        http.Handle("/", gorest.Handle())
        // start the actual web service
        http.ListenAndServ(":8787", nil)
    }

At runtime the HelloService will be bound. If tags are missing, or incorrect a panic will be thrown.

Service Customization

The rest service allows for the following tags to be specified:

    root            // The root of the service. URL part path string. Eg. "/tutorial/"
    produces        // The content type that the service will produce on requests.
                    //  Return types will be marshalled to that type. Eg. "application/json"
    consumes        // The type of content that the service will receive on requests.
                    //  Postdata will be unmarshalled from that type. Eg. "application/json"
    realm           // The security realm this service resides in. See Security section.
    allowGzip       // If by default all endpoints should allow their data to be returned as gzip data.
                    //  Can be overriden at each endpoint. Default is 'false' for backwards compatibility,

Each endpoint allows the following tags to be specified:

    method          // The HTTP method this endpoint handles. GET, POST, PUT, DELETE, HEAD
    path            // The path string this endpoint handles. Parameters can be encoding in this
                    //  string to be unmarshalled. Eg. "/car/{id:string}/{count:int}"
    output          // What this endpoint will return to the requester. Usually only specified on
                    //  on GET requests. Must be specified if the function returns something.
                    //  Eg. "[]MyStruct" or "string"
    role            // The security role of this endpoint inside the service. See Security section.
    produces        // Overrides the 'produces' on the Service. Allows an endpoint to specify its own
                    //  marshalling type.
    postdata        // The form/postdata type expected in the postdata. Only for POST/PUT requests.
                    //  Eg. "[]MyStruct" or "string"
    allowGzip       // If this endpoint should allow its data to be returned as gzip data.
                    //  If unset, it will use the service's value. Overrides service's value.

Security

Gorest allows for security to be handled by the user. It provides some helpers to deal with setup
and authorization of requests. On the service level we define a 'realm'. On an endpoint level we
define a 'role'. Using these two pieces of information authorization is verified.

First we need to define the authorizer for our realm. This will handle all requests for that realm,
and determine if the specified role and token are valid for the request.

    func SimpleAuthorizer(xsrfToken, role string, req *http.Request) (bool, bool, gorest.SessionData) {
        // xsrfToken is the token given to us by the request
        // role is the role specified by the endpoint being hit by the request
        // req is the request that is being made.

        // At this point you want to verify the authenticity of the xsrfToken
        // You should determine it is valid, or not, and also grab any
        // associated meta data you need from it.

        // Here we tell gorest that
        //  true: Yes this token is valid for this realm
        //  true: Yes this token is valid for this role
        //  SimpleSessionData{}: This is the session data associated with this token
        return true, true, SimpleSessionData{}
    }

SimpleSessionData is defined as:

    type SimpleSessionData struct {

    }
    // required method definition. You may want to store and return the xsrfToken here
    func (s SimpleSessionData) SessionId() string {
        return "1234"
    }

We then need to register our SimpleAuthorizer with gorest, before we register any services in that realm:

    gorest.RegisterRealmAuthorizer("simpleRealm", SimpleAuthorizer)

A service that defines a realm and roles:

    type HelloService struct {
        // defines the root of this service, and its meta data.
        gorest.RestService `root:"/tutorial/" realm:"simpleRealm"`
        // an end point that can be hit by a web request
        helloWorld      gorest.EndPoint `method:"GET" path:"/hello-world/" output:"string" role:"normalUser"`
    }
    func (h HelloService) HelloWorld() string {
        // get the session data
        sessionData := h.Session()
        // cast it to our type that returned in our authorizer
        // we can now access any methods/data in it.
        mySessionData := sessionData.(SimpleSessionData)
        return "Hello World"
    }

Marshallers

Custom Marshallers can be defined are registered. The following shows how a base64 marshaller could be 
implemented:

    func marshalBase64(v interface{}) (io.ReaderClose, error) {
        asString := fmt.Sprintf("%v", v)
        enc := base64.StdEncoding.EncodeToString(asString)
        return ioutil.NopCloser(bytes.NewBuffer(enc)), nil
    }
    func unmarshalBase64(data []byte, v interface{}) (error) {
        dec, err := base64.StdEncoding.DecodeString(string(data))
        if err != nil {
            return err
        }
        // psuedo-ish code.. may not actually compile
        switch v.(type) {
        case *string:
            *v = string(dec)
            break
        default:
            return errors.New("Bad unmarshall type")
        }
    }

    func main() {
        //...
        gorest.RegisterMarshaller("application/base64", &gorest.Marshaller{marshalBase64, unmarshalBase64})
        //...
    }

Logging

By default gorest uses the 'log' package to log. However, you can override this by implementing
the 'SimpleLogger' interface, and by calling the gorest.OverrideLogger(logger) method.

Recovering from Errors

If gorest or your code results in a runtime error a panic can be thrown. By default this is handled
and a 500 is returned. However, if you want to better manage this problem you specify a handler
by calling gorest.RegisterRecoveryHandler(handler).

Health

Some simple health information is reported via the health handler. See gorest.RegisterHealthHandler(handler).



*/
package gorest

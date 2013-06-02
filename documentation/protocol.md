# skynet client/server protocol

## Types

    ClientHandshake
    (defined in github.com/skynetservices/skynet ClientHandshake type)
    {
        
    }

    ServiceHandshake
    (defined in github.com/skynetservices/skynet ServiceHandshake type)
    {
        Registered bool
        ClientID string
    }

    RequestHeader
    (defined in net/rpc Request type)
    {
        ServiceMethod string
        Seq           uint64
    }

    ResponseHeader
    (defined in net/rpc Response type)
    {
        ServiceMethod string
        Seq           uint64
        Error         string
    }

    RequestInfo
    (defined in github.com/skynetservices/skynet RequestInfo type)
    {
        // OriginAddress is the reported address of the originating client, typically from outside the service cluster.
        OriginAddress string
        // RequestID is a unique ID for the current RPC request.
        RequestID  string
        // RetryCount indicates how many times this request has been tried before.
        RetryCount int
    }

    RequestIn
    (defined in github.com/skynetservices/skynet ServiceRPCIn type)
    {
        ClientID    string
        Method      string
        RequestInfo RequestInfo
        In          []byte
    }

    RequestOut
    (defined in github.com/skynetservices/skynet ServiceRPCOut type)
    {
        Out       []byte
        ErrString string
    }

## skynet protocol

1) Client/server handshake

Service: **ServiceHandshake**
* **Registered**: A value of false indicates the service will not respond to requests.
* **ClientID**: A UUID that must be provided will all requests.

Client: **ClientHandshake**

2) Client may begin sending requests. When done sending requests, the stream may be closed by the client.

Client: **RequestHeader**
* **ServiceMethod**: Use "**Name**.Forward", where **Name** is the service's reported name.
* **Seq**: Use a number unique to this session, usually by incrementing some counter.

Client: **RequestIn**
* **ClientID**: Must be the UUID provided by the **ServiceHandshake**.
* **Method**: The name of the RPC method desired.
* **RequestInfo**.**RequestID**: A UUID. If this is request is the direct result of another request, the UUID may be reused.
* **RequestInfo**.**OriginAddress**: If this request originated from another machine, that machine's address may be used. If left blank, the service will fill it in with the client's remote address.
* **In**: The BSON-encoded buffer representing the RPC's in parameter.

3) Service may synchronously send responses, in any order as long as the response corresponds to a request sent by the client. When the stream is closed by the client and all responses have been issued, the stream may be closed by the service.

Service: **ResponseHeader**
* **ServiceMethod**: Will be the same "**Name**.Forward" provided in the request.
* **Seq**: This number will match the **Seq** provided in the request to which this response corresponds.
* **Error**: Any rpc-level or skynet-level error. Empty string if no error. Errors in the actual service call are not put here.

Service: **RequestOut**
* **Out**: The BSON-encoded buffer represending the RPC's out parameter.
* **Error**: The text of the error returned by the service call, or the empty string if no error.

# skynet client/server protocol

## Types

    ClientHandshake
    (defined in github.com/bketelsen/skynet ClientHandshake type)
    {
        
    }

    ServiceHandshake
    (defined in github.com/bketelsen/skynet ServiceHandshake type)
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
    (defined in github.com/bketelsen/skynet RequestInfo type)
    {
        // OriginAddress is the reported address of the originating client, typically from outside the service cluster.
        OriginAddress string
        // RequestID is a unique ID for the current RPC request.
        RequestID  string
        // RetryCount indicates how many times this request has been tried before.
        RetryCount int
    }

    RequestIn
    (defined in github.com/bketelsen/skynet ServiceRPCIn type)
    {
        ClientID    string
        Method      string
        RequestInfo RequestInfo
        In          []byte
    }

    RequestOut
    (defined in github.com/bketelsen/skynet ServiceRPCOut type)
    {
        Out       []byte
        ErrString string
    }

## order of events

Using BSON as the encoding mechanism, the skynet protocol is as follows.

1) Service sends _ServiceHandshake_, using
* true for _Registered_ if the service is currently registered,
* and providing a new UUID for _ClientID_.

2) Client sends _ClientHandshake_, which as of this document's creation is empty.

3) Client may close the stream, ending the session, or go to step 4

4) Client sends _RequestHeader_, using 
* "_Name_.Forward" as the _ServiceMethod_ field, where _Name_ is the service's reported name,
* a number unique to this connection session for the _Seq_ field, possibly incremented for each request.

5) Client sends _RequestIn_, using
* the client ID received in step 1 for the _ClientID_ field,
* the service's method name for the _Method_ field,
* a unique value for the _RequestInfo_ fields's _RequestID_, unless the request is the result of an earlier request, in which case it may use the same request ID,
* an empty string for the _RequestInfo_'s _OriginAddress_ field, unless the request is proxied from another machine or is a result of a request from another machine, in which case it may be an address indicating the original source,
* the BSON-marshalled data to be decoded for the method's in-parameter for the _In_ field.

6) Service sends _ResponseHeader_, using
* the same _ServiceMethod_ as from step 4,
* the same _Seq_ as from step 4,
* an empty string for the _Error_, unless there was an rpc-level or skynet-level error, in which case it can contain the result of the error's .Error() method.

7) Service sends _RequestOut_, using
* the BSON-marshalled data encoded from the method's out-parameter for the _Out_ field,
* the string representation of any service-level error that occurred during the request, or an empty string if no error, for the _Error_ field.

go to step 3
Simple IPC protocol for unix domain sockets
===========================================

This library implements a very simple protocol for unix domain sockets to send
messages and file descriptors. Demos are included in demo/client and demo/server
where the client connects to the server via a socket file in the current
directory. The client gives the stdout file descriptor to the server which
writes a message to it.

Demo
----

The server:

    $ rm -f sock && go build ./demo/server && ./server
    2017/10/04 14:10:09 Receive client &{{0xc420020150}}
    2017/10/04 14:10:09 Received header simpleipc.Header{Seq:0x2a, Reserved0:0x0, Reserved1:0x0, NumFiles:0x1, Size:0x0, Files:[]*os.File{(*os.File)(0xc42000e050)}}

The client:

    $ go build ./demo/client && ./client
    2017/10/04 14:10:09 Connected to server &{{0xc4200200e0}}
    Hello world

Wire protocol
-------------

Messages are sent to the socket with a header and an optional payload. This
library is only specifying what the header format is and the payload is left to
the underlying application to specify.

Numbers are encoded in big endian.

Header format:

    0       8      16      24      32
    +-------------------------------+
    | Message number uint32 BE      |
    +-----------------------+-------+
    | Reserved              | NumF  |
    +-----------------------+-------+
    | Payload size uint32 BE        |
    +-------------------------------+

- message number is application specific. It is suggested that the client choose
  even sequence numbers and the server odd sequence numbers. When sending a
  message, the peer allocates a new sequence number in their pool, and when
  responding to a message, it uses the sequence number of the original request.
  It is also suggested that the first sequence numbers (0 and 1) are for
  messages with no reply.

- NumF: the number of file descriptors in ancillary data

- Payload size: the size of the following payload. It is suggested that long
  messages should be shared using a file descriptor instead with the apyload
  describing what is contained in the file descriptors and how they should be
  interpreted.

The header size is constant to ensure that it can be read with a single read and
that ancillary data attached with it are found correctly. See discussion about
what happens when messages with ancillary data are read partially:

https://unix.stackexchange.com/questions/185011/what-happens-with-unix-stream-ancillary-data-on-partial-reads

Contributions
-------------

The goal of this protocol is to stay simple, I don't plan to add features to it.
However if you want to contribute, you can contribute to tests and implementation
on other languages.

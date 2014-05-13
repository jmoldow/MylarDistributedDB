/*
 * socket_server.js - Server code for the distributed database.
 */

var net = Npm.require('net');
var fs = Npm.require('fs');
var server = net.createServer(function(conn) {
  conn.on('data', function(message) {
    console.log('message:');
    console.log(message.toString());
    // TODO Handle message. For now, just echo.
    conn.write(message.toString());
  });
  conn.on('close', function() {});
});

// File path of the Unix socket that the server should listen to.
var socketPath = '/var/tmp/824/';
try {
  fs.mkdirSync(socketPath);
} catch (e) { }
socketPath += 'meteor-' + getPort();

// Socket creation and cleanup code from
// <http://stackoverflow.com/questions/16178239/gracefully-shutdown-unix-socket-server-on-nodejs-running-under-forever/16502680#16502680>
server.on('error', function(e) {

  // If the server encounters an EADDRINUSE error when it starts to listen
  // to the socket, then one of two situations has occurred.  Either the
  // socket is still around from a previous use (but it is not curently in
  // use), or another server is currently running and is listening to the
  // socket. We can differentiate between these two scenarios by attempting
  // to connect to the socket as a client. In the former scenario, the
  // client connection will fail with an ECONNREFUSED, and then we know the
  // server can safely recover by unlinking the socket and trying again.
  if (e.code == 'EADDRINUSE') {
    var clientSocket = new net.Socket();

    // Handle error trying to talk to socket.  If the error is ECONNREFUSED,
    // no other server is listening, so the server can safely recover by
    // unlinking the socket and trying again.
    clientSocket.on('error', function(e) {
      if (e.code == 'ECONNREFUSED') {
        fs.unlinkSync(socketPath);
        server.listen(socketPath);
      }
    });

    // Attempt to talk to the socket as a client. If no error occurs, some
    // other server is running. Close the connection and do nothing else,
    // because recovery is impossible.
    clientSocket.connect(socketPath, function() {
      clientSocket.end();
    });
  }
});

server.listen(socketPath);


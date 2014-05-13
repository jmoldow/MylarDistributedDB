/*
 * socket_client.js - Server code for sending coordinator messages to the
 *                    distributed database.
 */

var net = Npm.require('net');

var socketPath = '/var/tmp/824/mmdb-in-' + getPort();

sendMessageToDB = function (message, callback) {
  if (!callback || !(callback instanceof Function)) {
    callback = function () { };
  }
  callback = Meteor.bindEnvironment(callback);

  var clientSocket = new net.Socket();

  clientSocket.on('error', function (e) {
    clientSocket.end();
    Meteor.setTimeout(function () {
      sendMessageToDB(message, callback);
    }, 2000);
  });

  clientSocket.on('data', function (reply) {
    clientSocket.end();
    callback(JSON.parse(reply));
  });

  clientSocket.connect(socketPath, function() {
    clientSocket.write(JSON.stringify(message));
  });
}

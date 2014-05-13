/*
 * socket_client.js - Server code for sending coordinator messages to the
 *                    distributed database.
 */

var net = Npm.require('net');

var socketPath = '/var/tmp/824/mmdb-in-' + getPort();

function _sendMessageToDB (message, callback) {
  var clientSocket = new net.Socket();

  clientSocket.on('error', Meteor.bindEnvironment(function (e) {
    clientSocket.end();
    setTimeout(Meteor.bindEnvironment(function () {
      sendMessageToDB(message, callback);
    }), 2000);
  }));

  clientSocket.on('data', Meteor.bindEnvironment(function (reply) {
    clientSocket.end();
    reply = reply.toString();
    console.log("go replied with");
    console.log(reply);
    try {
      reply = JSON.parse(reply);
    } catch (e) { reply = {}; }
    callback(undefined, reply);
  }));

  clientSocket.connect(socketPath, Meteor.bindEnvironment(function() {
    clientSocket.write(JSON.stringify(message));
  }));
};
_sendMessageToDB = Meteor._wrapAsync(_sendMessageToDB);

sendMessageToDB = function (message) {
  return _sendMessageToDB(message);
}

GetCoordinatorList = function (userId) {
  request = {
    Type: "LIST",
    Username: userId,
    Collection: "",
    Data: "",
    id: 0
  }
  reply = sendMessageToDB(request).List;
  return _.map(reply, getHostFromSocket);
}

CoordinatorPut = function (collectionName, userId, docId, doc) {
  request = {
    Type: "PUT",
    Username: userId,
    Collection: collectionName,
    Data: JSON.stringify(doc),
    id: docId
  }
  sendMessageToDB(request);
}

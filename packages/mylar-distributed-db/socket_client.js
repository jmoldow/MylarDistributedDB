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

nReplicas = 3;

GetCoordinatorList = function (userId) {
  request = {
    Type: "LIST",
    Username: userId,
    Collection: "",
    Data: "",
    id: 0
  }
  reply = JSON.parse(sendMessageToDB(request)).List;
  console.log("Replica list (unchanged): ");
  console.log(reply);
  reply = _.map(reply.slice(0, nReplicas), getHostFromSocket);
  console.log("changed:");
  console.log(reply);
  return reply;
}

CoordinatorPut = function (collectionName, userId, docId, doc) {
  request = {
    Type: "PUT",
    Username: userId,
    Collection: collectionName,
    Data: JSON.stringify(doc),
    id: docId
  }
  return sendMessageToDB(request);
}

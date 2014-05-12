Package.describe({
  summary: "A secure webmail system, built using Mylar, with end-to-end encryption and usable public key infrastructure"
});

// We use 'faye-websocket' for connections in server-to-server DDP, mostly
// because it's the same library used as a server in sockjs, and it's easiest to
// deal with a single websocket implementation.  (Plus, its maintainer is easy
// to work with on pull requests.)
Npm.depends({sockjs: "0.3.8", "faye-websocket": "0.7.2"});

Package.on_use(function (api) {
  api.use('livedata', ['client', 'server']);
  api.use(['check', 'random', 'ejson', 'json', 'underscore', 'deps',
           'logging', 'retry'],
          ['client', 'server']);

  // It is OK to use this package on a server architecture without making a
  // server (in order to do server-to-server DDP as a client). So these are only
  // included as weak dependencies.
  // XXX split this package into multiple packages or multiple slices instead
  api.use(['webapp', 'routepolicy'], 'server', {weak: true});

  // Detect whether or not the user wants us to audit argument checks.
  api.use(['audit-argument-checks'], 'server', {weak: true});

  // Allow us to detect 'autopublish', so we can print a warning if the user
  // runs Meteor.publish while it's loaded.
  api.use('autopublish', 'server', {weak: true});

  // If the facts package is loaded, publish some statistics.
  api.use('facts', 'server', {weak: true});

  api.use('callback-hook', 'server');

  api.use('accounts-base', 'server');

  api.use('principal', ['client', 'server']);

  // Transport
  api.use('reload', 'client', {weak: true});
  api.add_files('common.js');
  api.add_files('errors.js');
  api.export('CastErrorToCorrectSubclass', ['client', 'server']);
  api.add_files('livedata_connection.js', 'client');
  api.add_files('collection.js', ['client', 'server']);
  api.export('wrap_insert', ['client', 'server']);
  api.add_files('accounts_server.js', 'server');
  api.add_files('mylar_mail_client.js', 'client');
  api.add_files('mylar_mail_server.js', 'server');
  api.add_files('mylar_mail_common.js', ['client', 'server']);

  // we depend on LocalCollection._diffObjects, _applyChanges,
  // _idParse, _idStringify.
  api.use('minimongo', ['client', 'server']);
});

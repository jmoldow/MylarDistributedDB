var ffi = Npm.require('ffi')
var path = Npm.require('path')

var libfactorial = ffi.Library(path.join(path.resolve('.'), 'assets', 'packages', 'mylar-mail', 'libfactorial'), {
  'factorial': [ 'uint64', [ 'int' ] ]
})

var n = 6
for (var i = 0; i <= n; ++i) {
  var output = libfactorial.factorial(i);
  console.log('factorial(' + i + ') === ' + output);
}


client/compatibility
====================

Some JavaScript libraries only work when placed in the client/compatibility
subdirectory. Files in this directory are executed without being wrapped in
a new variable scope. This means that each top-level var defines a global
variable. In addition, these files are executed before other client-side
JavaScript files.

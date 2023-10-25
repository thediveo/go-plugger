/*
Package dyn discovers and loads .so Go plugins from the filesystem, so these
plugins then can register themselves with the plugger plugin mechanism.

# Important

The build tag/constraint “plugger_dynamic” must have been specified when using
this package; otherwise, [Discover] will panic as soon as it encounters any
plugin to be loaded dynamically.
*/
package dyn

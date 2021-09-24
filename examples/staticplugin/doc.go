/*

Package staticplugin demonstrates the usage of static plugins. It demonstrates
how to register a plugin function, as well as how to query the registered plugin
functions and calling them.

First, in your (static) plugins register your plugin function(s) or interface(s)
with a named plugin group.

Please note that in our example we don't explicitly specify a plugin group when
registering: plugger then automatically derives the plugin group from the name
of the parent(!) directory where the plugin package is located. For instance,
any plugin functions from "examples/static/plugin/plugins/foo" will be
registered with the plugin group "plugins", unless explicitly overriden.

Second, import your static plugins to ensure they are included in your build, so
they automatically register at startup. Importing plugins using the "_" blank
identifier usually suffices, as we don't need to access them specifically but
just want make sure that they're included in the build.

Third, query for the specific plugin group and then query that group for its
registered functions or interfaces (symbols). Please see the example below.

*/
package staticplugin

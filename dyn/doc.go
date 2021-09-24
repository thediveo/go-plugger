//go:build plugger_dynamic
// +build plugger_dynamic

/*
Package dyn discovers and loads .so Go plugins from the filesystem, so these
plugins then can register themselves with the plugger plugin mechanism.
*/
package dyn

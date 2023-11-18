# See also: ramls/Makefile (used only for validation and documentation)

**default**: target/ModuleDescriptor.json target/mod-reporting

target/ModuleDescriptor.json:
	(cd target; make)

target/mod-reporting:
	(cd src; make)

clean:
	(cd target; make clean)



COOKBOOK=mongodb
BRANCH=master

BUILD_DIR=../build
DIST_PREFIX=$(BUILD_DIR)/$(COOKBOOK)

all: metadata.json

clean:
	-rm metadata.json

metadata.json:
	-rm $@
	knife cookbook metadata -o .. $(COOKBOOK)
	
dist: clean metadata.json
	mkdir -p $(BUILD_DIR)
	version=`python -c "import json;c = json.load(open('metadata.json')); print c.get('version', 'UNKNOWN')"`; \
	tar --exclude-vcs --exclude=Makefile -cvzf $(DIST_PREFIX)-$$version.tar.gz ../$(COOKBOOK)

SUBDIRS := golang java

.PHONY: all
all: $(SUBDIRS)

.PHONY: docker.build
docker.build: $(SUBDIRS)

.PHONY: $(SUBDIRS)
$(SUBDIRS):
	$(MAKE) -C $@ $(MAKECMDGOALS)

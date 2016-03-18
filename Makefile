IMAGE_NAME = service

build:
	sh -c "'$(CURDIR)/script/build.sh'"
	echo "==> Create docker dmp..."
	docker build -t $(IMAGE_NAME) .

dry-build:
	sh -c "'$(CURDIR)/script/build.sh'"
	echo "==> Create docker dmp..."
	@CID=$$(docker create $(IMAGE_NAME)) && \
		docker cp $(CURDIR)/bin/service $$CID:/usr/bin/service && \
		docker cp $(CURDIR)/../dmp/bin/dmp $$CID:/usr/bin/dmp&& \
		docker cp $(CURDIR)/supervisord.conf $$CID:/etc/supervisor/conf.d/supervisord.conf && \
		docker stop $$CID && \
		docker commit $$CID $(IMAGE_NAME) && \
		docker rm -vf $$CID

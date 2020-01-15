run:
	@echo "=============starting server============="
	docker-compose -f docker-compose.dev.yml up --build

test:
	@echo "=============running test============="
	docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit && docker-compose -f docker-compose.test.yml down

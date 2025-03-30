.PHONY: dev
dev:
	@watchexec -r -e go --wrap-process session -- "go run ."

.PHONY: log
log:
	@tail -f debug.log

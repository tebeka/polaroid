test:
	go test -race ./...
	go tool staticcheck ./...
	go tool govulncheck ./...

release-patch: check-clean
	$(MAKE) release version=$$(go tool svu patch)

release-minor: check-clean
	$(MAKE) release version=$$(go tool svu minor)

release: check-clean test
	git tag -a $(version) -m "release $(version)"
	git push origin $(version)

check-clean:
	@test -z "$$(git status --porcelain)" || (echo "error: worktree is not clean"; git status --short; exit 1)

.PHONY: test release release-patch release-minor check-clean

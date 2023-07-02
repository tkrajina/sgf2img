.PHONY: run
run: clean
	go run cmd/sgf2img/*go -g -c sgf/*sgf
	mkdir -p examples/bw
	mv -v sgf/*png examples/bw

	go run cmd/sgf2img/*go -c sgf/*sgf
	mkdir -p examples/color
	mv -v sgf/*png examples/color

	go run cmd/sgf2img/*go -c -t svg sgf/*sgf
	mkdir -p examples/svg
	mv -v sgf/*svg examples/svg

.PHONY: clean
clean:
	-rm sgf/*png
	-rm sgf/*svg
	-rm sgf/*csv
	-rm -Rf examples/*

.PHONY: install
install:
	@if [ -z "$(DIR)" ]; then \
	    echo 'target DIR missing'; exit 1; \
	fi
	go build -o $(DIR)/sgf2img ./cmd/sgf2img/*go
	go build -o $(DIR)/sgfrename ./cmd/sgfrename/*go
	go build -o $(DIR)/sgfinfo ./cmd/sgfinfo/*go
	go build -o $(DIR)/sgffindpos ./cmd/sgffindpos/*go
	go build -o $(DIR)/sgflongestmainline ./cmd/sgflongestmainline/*go
	go build -o $(DIR)/sgf2ankicsv ./cmd/sgf2ankicsv/*go
	go build -o $(DIR)/sgfs2md ./cmd/sgfs2md/*go
	go build -o $(DIR)/sgfcleankatrain ./cmd/sgfcleankatrain/*go
	go build -o $(DIR)/sgfcleancomments ./cmd/sgfcleancomments/*go

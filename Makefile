# Copyright 2018 The Fractal Team Authors
# This file is part of the fractal project.
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program. If not, see <http://www.gnu.org/licenses/>.

TEST = $(shell go list ./... | grep -v test | grep -v txpool)

all:
	@go install ./cmd/ft
	go build ./cmd/ft
	mv ft ./build/bin

	@go install ./cmd/ftkey
	go build ./cmd/ftkey
	mv ftkey ./build/bin
	
run:
	@./build/bin/ft
stop:
clear:
test: all
	@echo $(TEST)
	go test $(TEST)

.PHONY: all run stop clear test

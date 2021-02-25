CC = gcc
CFLAGS = `pkg-config --cflags --libs glib-2.0`

EXECUTABLE = main
SOURCES = main.c util.c sha256.c
HEADERS = main.h sha256.h

all:
	$(CC) -o $(EXECUTABLE) $(SOURCES) $(HEADERS) $(CFLAGS)

clean:
	rm -f $(EXECUTABLE)

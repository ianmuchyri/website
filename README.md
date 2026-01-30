# Abstract Machines Website

This repository contains the source code for the [Abstract Machines](https://absmach.eu) website and blog.

## Project Structure

- `content/blogs/`: Markdown files for blog posts.
- `img/blogs/`: Images used in blog posts.
- `scripts/`: The Go-based static site generator.
- `scripts/templates/`: HTML templates for the blog listing and individual posts.
- `blog/`: Generated static files (do not edit manually).
- `index.html`: The main landing page.

## Prerequisites

- **Go**: Required to run the blog builder.
- **Make**: Used for task automation.

## Guidelines for Contributors

To add a new blog post, follow these steps:

1. Create your content in `content/blogs/` (see [WRITING.md](WRITING.md)).
2. Build the site locally to generate the static files:
   ```bash
   make clean && make build
   ```
3. Commit both the source Markdown files **and** the generated files in the `blog/` folder.
4. Open a Pull Request.

## Running locally
To run the built website locally, first build it with:
```bash
make build
```
Then, you can run it with:
```bash
make serve
```
Open your browser at http://localhost:8080.
If you want to change the port (in case 8080 is already taken), you can run:
```bash
PORT=8081 make serve
```

## Documentation

- [How to Write a Blog Post](WRITING.md)

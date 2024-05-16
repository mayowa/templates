# Templates

This library is a lightweight templating package designed for use with Echo(link) but which can also be used standalone (Can it? Well duh. Do you use Echo in the tests? It just extends the stdlib, as long as you call Render it'll do its job).

# Why?

Templates has two primary goals:

- Make your templates easy to write
- Make your templates easy to read by abstracting away repetitive markup
  Templates makes your life easy without being complex! You can get psuedo-inheritance like this:

```go
tmpl := template.New(root, extension)

// render the index template usinge the "default" layout
tmpl.Render("default","index")
```

If your default template looks like this:

```tmpl
// default.tmpl
Hello {{ block "content" . }} {{ end }}!
```

And your index template looks like this:

```tmpl
//index.tmpl
{{ define "content" }} World {{ end }}
```

Then templates will render output that looks like this:

```tmpl
//output
Hello World!
```

Templates also allows you to reuse template code by creating components like this:

TODO

# Getting started

Using this thing is pretty simple!

```go
// Create a template struct  by passing in the root folder and extension for all templates
create tmpl := template.New(root, extension)

// Render a template through an io.Writer with a layout and context
tmpl.Render(io.Writer, "template layout", "template name", data)

// Render a template without a layout
tmpl.Render(io.Writer, "", "template name", data)

// render all templates in the settings folder
tmpl.Render(io.Writer, "layout","settings")

```

## Things to note

- Any templates inside a `templateRoot/shared` folder will be parsed on every call to Render
- Don't add the file extension to template names when calling Render or referencing one from a template
- A string can be rendered as a template like any other
- Folders (other than those in shared) can treated as a group of templates.
- If a template with the same name as the folder exists in the folder, it is parsed first

# Example folder structure

```
- <template root folder>
    default.tmpl
    - shared
        button.tmpl
        toolbar.tmpl
    - index.tmpl
    - about.tmpl
    - settings
        settings.tmpl
        options.tmpl
        form.tmpl
    - veggies
        carrots.tmpl
        spinach.tmpl
        cabbage.tmpl
```

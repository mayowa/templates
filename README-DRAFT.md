# Templates

This library is a lightweight package that adds convenience features to the standard library's template system.

# Why?

Templates has two primary goals:

- Make your templates easy to write
- Make your templates easy to read by abstracting away repetitive markup
- Templates makes your life easy without being complex! You can render a template like this:

```go
// create a template struct
tmpl := template.New(root, extension)

// render the index template
tmpl.Render(io.Writer, "index", data)
```

This package provides pseudo - inheritance capabilities. By placing an "extends" directive in a comment at the top of your template file, you can specify another template as a layout. Here's an example. If your index template looks like this:

```tmpl
// index.tmpl
{{/* extends default */  }}
{{ define "content" . }} World {{ end }}
```

And your default template looks like this:

```tmpl
//default.tmpl
Hello {{ block "content" . }} {{ end }}!
```

Then calling `tmpl.Render` will output something like this:

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

- When Render is parsing the templates passed to it, all templates inside the `templateRoot/shared` folder will be parsed alongside the templates you pass.
- Don't add the file extension to template names when calling Render or referencing one from a template
- A string can be rendered as a template like any other
- If you pass the name of a folder to Render, all templates inside that folder are parsed as a group. If a file `folderName/folderName.ext` exists, it will be parsed first and treated as the layout for all other files inside `folderName`.
- Need to work out folders
- Need to edit the examples to reflect the new render/layout method
- Components

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

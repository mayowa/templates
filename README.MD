# Templates
a wrapper around html/templates

# example folder structure
```
- <template root folder>
    default.tmpl
    - shared
        button.tmpl
        toolbar.tmpl
    - index.tmpl
    - settings
        settings.tmpl
        options.tmpl
        form.tmpl
```

## core ideas
shared templates should be in the shared folder
template names will not include file extension when passed to Render or referenced within other templates 
a template can be rendered without a layout
a string can be rendered as a template and make use of layout and shared templates

folders (other than those in shared) can treated as a group of templates. if a template with the same name 
as the folder exists in the folder, it is parsed first
```go 
templates.Render("default","index")

// render all templates in the settings folder 
templates.Render("default","settings")
```

## Who is this library for?

Golang developers that build applications where the UI is rendered from the server using html/template

## What does the library do?

Templates makes writing and using template partials convenient, because partials make it easier to write HTML that's easy to read and reuse.

## How does it do that?

In Go, using partials normally works like this: you have a `main` template and a `partial` template. You invoke the partial inside another template like this:

```go
// main.tmpl
{{ template "partial1" data }}

{{ template "partial2" data }}

{{ template "partial3" data }}

```

And then when you want to render `main`, you have to parse each you use partial with it:

```go
  //handler.go
  mainTmpl.Parse("partial1", "partial2", "partial3").Execute(io.Writer)
```

This get the job done, but is inconvenient. Using templates allows you to do this instead:

```go
//main.tmpl
<Partial1 data="data" />

<Partial1 data="data" />

<Partial1 data="data" />

```

And to render main like this:

```go
  tmpl.Render("main")
```

Which as i'm sure you'll agree, is easier to read. Now that you know the what, why and how, let's get into the details of using templates.

## Basic Usage

To do anything with Templates, you need to call the `New()` function and pass it the root folder where your Go templates live to get a `templates.template` object (confusing naming, I know). The `template` object doesn't represent an individual template though: it doesn't represent any of them. It's a struct with interesting properties we'll get to later, but its main job is to allow you to parse a Go template, execute it with some data, and then write the output with an io.Writer.

To achieve this, the `templates.template` object has a `Render` function that takes an `io.Writer`, the name of the template file (minus) the extension, and the data you want to use to execute the template. So here's the simplest way you can use templates: Create the main template like this:

```go
//./templates/main.tmpl

This is the main template
```

```go
//handler.go

func handleMain(r Request, w http.ResponseWriter) {
  tmpl := templates.New("./templates")
  tmpl.Render(w, "main", nil)
}
```

Now, let's assume you want to use a partial inside main to greet your user. Create a file in `./templates/components` called `hello.tmpl`. Assume hello looks something like this:

```
Hello {{ .name }}!
```

You can use now hello inside main like this:

```go
//./templates/main.tmpl

This is the main template

<Hello name="{{ .Name }}" />
```

And render main like this:

```go
//handler.go

func handleMain(r Request, w http.ResponseWriter) {
  tmpl := templates.New("./templates")
  data = make(map[string]any, 0)
  data["Name"] = "Bob"
  tmpl.Render(w, "main", data)
}
```

And that's all you need to get started. The following sections describe the various features of the library in more detail.

## TemplateOptions

## Components in Depth

As you might have noticed, the syntax for partials is meant to mimic html tags, but...how does that even work? The templates themselves are regular Go. What the Templates library does is pre-process the templates to convert the HTML - esque syntax into a call to Go's `template` action where the data passed is a map of the attributes and values. Each template tag is a separate call. Using our previous example, the main template:

```go
//./templates/main.tmpl

This is the main template

<Hello name="{{ .Name }}" />
```

Will be converted to something like this:

```go
//./templates/main.tmpl

This is the main template


{{ template "hello.tmpl" (map "name" {{ .Name }} "_isSelfClosing" true "_isEnd" false) }}
```

Components can also be used in open-and-close tag pairs, so a component call like this:

```go
<Message>
  Boo yeah!
</Message>
```

Will be converted to this:

```go
{{ template "message.tmpl" (map "_isSelfClosing" false "_isEnd" false)  }}
  Boo yeah
{{ template "message.tmpl" (map "_isSelfClosing" false "_isEnd" true)  }}

```

the map function takes a list of space-separated key-value pairs and passes a map to the template. Because components can either be used as self-closing or paired, with opening and closing tags the library adds a couple more variables to the map so that you can do conditional rendering in your partial to render different content for open, close, and self closing tags. Here's an example:

```go
<Message
```

## Layouts

## Shared

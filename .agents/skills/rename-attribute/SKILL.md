---
name: rename-attribute
description: |
  Renames a Terraform attribute safely without changing its schema or producing an
  unnecessary diff, while enabling users to use both the previous and subsequent
  attribute names for a deprecation period.
---

# Rename Terraform Attribute

## Overview

A user may ask you to rename a Terraform attribute. A Terraform attribute has
a name and a schema representing the types of data the attribute can contain.
Renaming a Terraform attribute keeps the original attribute and creates a new
attribute with the same schema, introducing a deprecation period so the user can
migrate from the old attribute to the new attribute.

## Workflow

### Determine Attribute Schema

Use "Determine Attribute Schema" section.

### Create New Attribute

Use "Create New Attribute" section.

### Modify New and Old Attribute Schemas

Use "Modify New and Old Attribute Schemas" section.

### Modify CRUD Methods

Use "Modify CRUD Methods" section.

## Determine Attribute Schema

Find the `Schema` method on the resource and look for the attribute the user
asked to be renamed. Take note of the schema for the attribute and whether the
attribute is `Required`, `Optional`, or `Computed`. This is the schema that will
need to be used for the new attribute.

## Create New Attribute

Create a new attribute on the resource within the `Schema` method using the new
name that the user told you. The schema for this attribute must be identical to
the schema for the old attribute. The only difference should be the attribute
name.

## Modify New and Old Attribute Schemas

### Struct Fields

There's a struct type that represents the resource that Terraform uses to
serialize and deserialize Go types to Terraform types and vice versa. The type
is usually something with "model" in the name. You should be able to find the
field by looking for a `tfsdk` struct field tag with the value that matches the
name of the attribute from `Schema`.

Once you find this struct, rename the field representing the old attribute and
append `Deprecated`. Create a new field representing the new attribute with a
CamelCase name based on the name the user gave you (e.g., `NewAttribute`).

### Schema Definition

Update the schema for the both the new and old attributes with the following
changes.

* Remove `Required`.
* Add `Optional: true`.
* Add `Computed: true`.

### Deprecation Message

Add `DeprecationMessage` to the old attribute with a message that tells the user
to use the new attribute. Update the `Description` and/or `MarkdownDescription`
for the old attribute to note that it's deprecated.

### Validation

Add a `validator.ExactlyOneOf` validator to the old attribute with a path
expression matching the new attribute. There should be no such validator on the
new attribute.

## Modify CRUD Methods

### Create

At the top of the `Create` method there will be logic that loads the plan data
into a variable. Find the variable name so we can use it for future operations.

```go
var data ExampleModel
resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
```

Set values for the new and old attribute values in the data variable that was
retrieved earlier. The values will likely be retrieved from an earlier API
operation to create the resource. It must look something like this.

```go
data.NewAttribute = // ... 
data.OldAttribute = // ...
```

Anywhere in `Create` that the old attribute was used, replace that usage with
the following conditional logic. Change `any` to be the correct type for the
attribute and use the correct `Value` method.

```go
	var attribute any
	if !data.Hostname.IsNull() && !data.Hostname.IsUnknown() {
		attribute = data.NewAttribute.Value()
	} else {
		attribute = data.OldAttribute.Value()
	}
```

### Read

At the top of the `Read` method there will be logic that loads the state data
into a variable. Find the variable name so we can use it for future operations.

```go
var data ExampleModel
resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
```

Set values for the new and old attribute values in the data variable that was
retrieved earlier. The values will likely be retrieved from an earlier API
operation to read the resource. It must look something like this.

```go
data.NewAttribute = // ... 
data.OldAttribute = // ...
```

### Update

At the top of the `Update` method there will be logic that loads the plan data
into a variable. Find the variable name so we can use it for future operations.

```go
var data ExampleModel
resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
```

Set values for the new and old attribute values in the data variable that was
retrieved earlier. The values will likely be retrieved from an earlier API
operation to update the resource. It must look something like this.

```go
data.NewAttribute = // ... 
data.OldAttribute = // ...
```

Anywhere in `Update` that the old attribute was used, replace that usage with
the following conditional logic. Change `any` to be the correct type for the
attribute and use the correct `Value` method.

```go
	var attribute any
	if !data.Hostname.IsNull() && !data.Hostname.IsUnknown() {
		attribute = data.NewAttribute.Value()
	} else {
		attribute = data.OldAttribute.Value()
	}
```

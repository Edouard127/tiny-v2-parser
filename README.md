The Tiny V2 mapping format is a structured text format used for obfuscation and deobfuscation, mainly used in Fabric

## Tiny V2 Format

For more information, consult the [official format documentation](https://wiki.fabricmc.net/documentation:tiny2) from Fabric

### Header Line
    Format: `tiny\t<major>\t<minor>\t<namespace1>\t<namespace2>...`
    Example: tiny\t2\t0\tofficial\tnamed defines major version 2, minor version 0, with two namespaces: "official" and "named".

#### Properties (optional)
    Indented lines with 1 tab, e.g., \t<key>\t<value>.
    Special property: escaped-names (no value) indicates names use escape sequences.

### Hierarchy & Indentation
Lines are indented with tabs to denote nesting:
- 0 tabs: Top-level classes.
- 1 tab: Fields (f), methods (m), or comments (c) under a class.
- 2 tabs: Parameters (p), local variables (v), or comments under methods.
- 3 tabs: Comments under parameters or variables.

### Line Types
Each line starts with a type identifier after indentation:

#### Class (c)

    Format: c\t<name-namespace1>\t<name-namespace2>...
    Example: c\tnet/minecraft/Class\tNamedClass maps the class across namespaces.

#### Field (f)

    Format: \tf\t<descriptor>\t<name-namespace1>\t<name-namespace2>...
    Example: \tf\tI\tfield_123\tnamedField defines a field with descriptor I.

#### Method (m)

    Format: \tm\t<descriptor>\t<name-namespace1>\t<name-namespace2>...

    Example: \tm\t()V\tmethod_456\tnamedMethod defines a method with descriptor ()V.

#### Parameter (p)

    Format: \t\tp\t<index>\t<name-namespace1>...

    Example: \t\tp\t0\tparam defines a parameter at index 0.

#### Local Variable (v)

    Format: \t\tv\t<index>\t<start-offset>\t<lvt-index>\t<name-namespace1>...

    Example: \t\tv\t1\t0\t2\tlocalVar defines a local variable.

#### Comment (c at non-zero indent)

    Format: \t...c\t<comment-text>

    Example: \tc\tThis is a class comment adds a comment to a class.

Comment will still be parsed if they don't have a proper identifier or indentation, this a feature or
the parser, not a bug

### Escaping

   Enabled by the escaped-names property.

   Escape Sequences:

        \\ → \
        \n → Newline
        \r → Carriage Return
        \0 → Null
        \t → Tab

# Translating CASCI

## Current status

The application currently ships English strings only. The locale directory and
translator workflow required by AI.md have not been implemented yet, so there
is not currently a supported way to add a language through the application.
This page records the project rules to use when localization is implemented;
it does not claim that translations are active today.

## Planned file and registry workflow

When the localization system lands, English will be the source locale.
Translators should work in one locale JSON file and preserve its keys and
structure. The planned file location is
`src/common/i18n/locales/{language-code}.json`; this directory does not exist
in the current source tree.

For a future locale contribution:

1. Copy the English source locale to a lowercase BCP 47 language-code file.
2. Translate every value; keep keys unchanged and fill language name, native
   name, text direction, version, and translator credit metadata.
3. Register the language and its direction and plural categories in the
   language registry and configuration.
4. Preserve every placeholder exactly, including braces and names.
5. Keep product names, protocol names, code, paths, config keys, URLs, and
   literal values users enter untranslated.
6. Use complete phrases and a consistent register; do not concatenate
   translated fragments.
7. Check plural categories against CLDR for that language.
8. Review layout at 320, 768, and 1280 CSS pixels; check right-to-left layout
   when applicable.
9. Add a font-stack entry when a script may not be covered by the system font
   stack.
10. Run the localization tests, update the README language list, and add the
    translator credit here and in locale metadata.

## Glossary

Use these terms consistently when translating the interface. The glossary is
provisional until the project adds a translation registry and glossary file.

| English term | Meaning |
|--------------|---------|
| Token | An authentication credential presented to the API |
| Build | One execution of a project's pipeline |
| Node | A machine registered to execute build work |
| Revoke | Invalidate an existing credential or access grant |
| Preferences | Browser-level appearance and language settings |

## Review and credit

Maintainers should check placeholders, plural categories, glossary terms, and
layout before merging a translation. Human review can refine machine
translations; native-speaker review is welcome but is not currently a project
gate. List translator names or handles here only after a contribution is
accepted.

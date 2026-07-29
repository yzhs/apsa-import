def type(name):
    select(."@type" == name);

."@graph" | map({key: ."@type", value: .}) | from_entries |
(.Article.keywords // if .Recipe.keywords then [.Recipe.keywords | split(",")[] | trim] else [] end) as $tags |
.Recipe |
{
    title: .name,
    source: $source,
    tags: $tags,
    yield: .recipeYield,
    time: {preparation: .prepTime, cooking: .cookTime, total: .totalTime},
    steps: [
        {
            ingredients: .recipeIngredient,
            instructions: ([.recipeInstructions[] | (type("HowToSection").itemListElement[] // .) | type("HowToStep").text | select(. != "")] | join("\n\n"))
        }
    ]
} |
del(.. | nulls)

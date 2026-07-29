def type(name):
    select(."@type" == name);

."@graph"[] |
(type("Article").keywords // if type("Recipe").keywords then [.Recipe.keywords | split(",")[] | trim] else [] end) as $tags |
type("Recipe") |
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

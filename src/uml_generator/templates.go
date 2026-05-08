package uml_generator

// systemPromptTemplate is the format reference injected into every AI generation call.
// It mirrors the content of template.md at the project root.
// Keeping it here (compiled in) avoids runtime file I/O and embed path issues.
//
// To update the prompt, edit this constant — it is the single source of truth
// for the Mermaid syntax rules used by the AI.
const systemPromptTemplate = `
# UML Solution File Format Guide

## Output Rules
- Return ONLY the raw Mermaid code block (` + "```" + `mermaid ... ` + "```" + `), no explanation.
- Every element MUST have a score tag __1__ (default score = 1).
- Use ~T~ instead of <T> for generics (Mermaid limitation).
- Use | to express polymorphism / alternative names or types.

## Syntax Reference

### Class definition
` + "```" + `
class ClassName {
  <<Stereotype>>
  [visibility] "[name|altName]" : "[Type1|Type2]" [Modifier] __score__
  [visibility] "[methodName|altName]([param: Type1|Type2])" "[ReturnType1|ReturnType2]" [Modifier] __score__
}
` + "```" + `

- Stereotype: <<Abstract>>, <<Interface>>, <<Enum>>, etc. (optional)
- Visibility: + public · - private · # protected · ~ package
- Modifier (optional): {ReadOnly}, {Static}, {Abstract}
- Score tag: __d__ where d is the point value (e.g. __1__, __0.5__)

### Relationships
ClassA <|-- ClassB : __1__   (Inheritance)
ClassA ..|> ClassB : __1__   (Realization)
ClassA o-- ClassB  : __1__   (Aggregation)
ClassA *-- ClassB  : __1__   (Composition)
ClassA --> ClassB  : __1__   (Association)
ClassA ..> ClassB  : __1__   (Dependency)

### Polymorphism with |
- "getName|fetchName"           — either method name is accepted
- "String|char[]|CharSequence"  — either type is accepted
- "getId(u: User|Member)"       — either param type is accepted

## Full Example
` + "```" + `mermaid
classDiagram
    Animal <|-- Duck : __1__

    class Animal {
      <<Abstract>>
      + "name|animalName" : "String|char[]" __1__
      + "getName|fetchName()" "String" Abstract __1__
    }

    class Duck {
      - "beakColor|beakColour" : "String|Color" __1__
      + "swim|move()" "void" __1__
      - "canEat(prey: Animal|Object)" "bool|boolean" __1__
    }

` + "```"

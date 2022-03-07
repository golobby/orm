# API Documentation
If you prefer (like myself) a more api oriented document this part is for you. Almost all functionalities of ORM is exposed thorough
simple functions of ORM, there are 2 or 3 types you need to know about:
- `Entity`: Interface which all structs that are database entities should implement, it only has one method that configures entity for the ORM.

Now let's talk about ORM functions. also, please note that since Go1.18 is on the horizon we are using generic feature extensively to give
a really nice type-safe api.


- Basic CRUD APIs
    - `Insert` => Inserts given `Entity`.
    - `InsertAll` => batch insert on a table
    - `Find` => Finds record based on given PK and Type parameter.
    - `Save` => Upserts given entity.
    - `Update` => Updates given entity.
    - `Delete` => Deletes given entity from database.

*Note*: for relationship we try to explain them using post/comment/category sample.

- Relationships
    - `Add` [TBA]: This is a relation function, inserts `items` into database and also creates necessary wiring of relationships based on `relationType`.
    - `BelongsTo`: This defines a hasMany inverse relationship, relationship of a `Comment -> Post`, each comment belongs to a post.
    - `BelongsToMany`: Relationship of `Post <-> Category`, each `Post` has categories and each `Category` has posts.
    - `HasMany`: Relationship of `Post -> Comment`, each post has many comments.
    - `HasOne`:


- Custom and Raw queries
    - `Exec`
    - `ExecRaw`
    - `Query`
    - `RawQuery`


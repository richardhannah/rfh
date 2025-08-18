# Questions about the `add` command implementation

## 1. What should the `add` command do exactly?
- Download a package from the registry and extract it to the current workspace?
- Should it create a local installation directory structure?
- Should it update some kind of lockfile or dependency manifest?

## 2. Where should downloaded packages be stored?
- In the current working directory?
- In a specific subdirectory (like `node_modules`, `.rfh/packages`, etc.)?
- Should it respect any existing project structure?

## 3. What should happen if a package is already installed?
- Overwrite, skip, or prompt the user?
- Should it support version resolution (latest, specific versions)?

## 4. Integration with the API
- I see there are `getPackageHandler` and `downloadBlobHandler` endpoints - should `add` use these?
- Does the package download process need authentication?

## 5. Expected behavior
- Should it be similar to `npm install`, `cargo add`, or something else?
- Should it install dependencies recursively or just the specified package?

## Additional context
Looking at the search command, I can see it uses the client to call API endpoints. The `add` command would likely need to:
1. Parse package name and version from the argument
2. Call the API to get package metadata 
3. Download the package archive
4. Extract it to the appropriate location
# Dynamic Form Engine (Expo + React Native)

Schema-driven form renderer with typed validation, conditional visibility, runtime custom field registration, and explicit form state transitions.

## Features

- Render forms from JSON schema
- Built-in field types: `text`, `select`, `date`, `multi-select`
- Composable field validation:
  - `required`
  - `minLength`
  - `maxLength`
  - `regex`
  - `custom` validators (inline function or registry key)
- Conditional field visibility (`visibleWhen`)
- Runtime registration of custom field components
- Form status lifecycle:
  - `pristine -> dirty -> validating -> submitting -> success/error`
- TypeScript-first contracts for schema and state

## Project Structure

- `App.tsx` - app entry and sample form usage
- `src/types/schema.ts` - typed schema, validation rules, and form status
- `src/engine/visibility.ts` - conditional visibility logic
- `src/engine/validation.ts` - validation engine + validator registry support
- `src/engine/formController.ts` - pure form state controller
- `src/hooks/useDynamicForm.ts` - React adapter for controller
- `src/fields/FieldRegistry.tsx` - runtime field registry APIs
- `src/fields/defaultFields.tsx` - default field implementations
- `src/components/FormRenderer.tsx` - schema renderer
- `src/sampleSchema.ts` - working schema and sample custom validator
- `src/engine/formController.test.ts` - unit tests for core behavior

## Getting Started

### 1) Install dependencies

```bash
npm install
```

### 2) Run the app

```bash
npm start
```

Optional:

```bash
npm run android
npm run ios
npm run web
```

### 3) Run tests

```bash
npm test
```

Watch mode:

```bash
npm run test:watch
```

## Schema Example

```ts
import { FormSchema } from "./src/types/schema";

const schema: FormSchema = {
  id: "profile-form",
  title: "Profile Setup",
  fields: [
    {
      id: "name",
      label: "Name",
      type: "text",
      validation: [{ type: "required" }, { type: "minLength", value: 3 }],
    },
    {
      id: "role",
      label: "Role",
      type: "select",
      options: [
        { label: "Developer", value: "dev" },
        { label: "Manager", value: "manager" },
      ],
      validation: [{ type: "required" }],
    },
    {
      id: "startDate",
      label: "Start Date",
      type: "date",
      visibleWhen: { field: "role", equals: "manager" },
      validation: [{ type: "required" }],
    },
  ],
};
```

## Using the Renderer

```tsx
import { FormRenderer } from "./src/components/FormRenderer";
import { defaultSchema, sampleCustomValidators } from "./src/sampleSchema";

<FormRenderer
  schema={defaultSchema}
  customValidators={sampleCustomValidators}
  onSubmit={async (values) => {
    // submit logic
    return values;
  }}
/>;
```

## Registering a Custom Field Type at Runtime

```tsx
import { TextInput } from "react-native";
import { FieldComponentProps } from "./src/fields/FieldRegistry";

const CurrencyField = ({ value, onChange }: FieldComponentProps) => (
  <TextInput
    value={typeof value === "string" ? value : ""}
    onChangeText={(text) => onChange(text)}
    keyboardType="numeric"
    placeholder="0.00"
  />
);

<FormRenderer
  schema={schema}
  customFields={[{ type: "currency", component: CurrencyField }]}
  onSubmit={async (values) => values}
/>;
```

## Validation Notes

- Hidden fields are not validated.
- Custom validators can be:
  - inline function in schema, or
  - string key resolved from `customValidators`.
- Missing custom validator keys return explicit validation errors.

## Design Goals

- Keep core logic loosely coupled from UI
- Make validation and transitions testable in isolation
- Support extensibility without editing engine internals
- Keep behavior explicit and predictable for live walkthrough discussions

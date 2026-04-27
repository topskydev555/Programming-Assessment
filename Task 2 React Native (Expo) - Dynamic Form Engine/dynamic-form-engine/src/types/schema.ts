export type FieldType = "text" | "select" | "date" | "multi-select" | string;

export type PrimitiveFieldValue = string | number | boolean | null;
export type FieldValue = PrimitiveFieldValue | PrimitiveFieldValue[];
export type FormValues = Record<string, FieldValue>;

export type ValidationError = {
  fieldId: string;
  message: string;
};

export type ValidatorFn = (
  value: FieldValue | undefined,
  values: FormValues
) => string | null;

export type ValidationRule =
  | { type: "required"; message?: string }
  | { type: "minLength"; value: number; message?: string }
  | { type: "maxLength"; value: number; message?: string }
  | { type: "regex"; value: string; flags?: string; message?: string }
  | { type: "custom"; validator: ValidatorFn | string; message?: string };

export type VisibilityCondition = {
  field: string;
  equals: PrimitiveFieldValue;
};

export type SelectOption = {
  label: string;
  value: PrimitiveFieldValue;
};

export type FieldSchema = {
  id: string;
  label: string;
  type: FieldType;
  placeholder?: string;
  options?: SelectOption[];
  validation?: ValidationRule[];
  visibleWhen?: VisibilityCondition;
};

export type FormSchema = {
  id: string;
  title?: string;
  fields: FieldSchema[];
};

export type FormStatus =
  | "pristine"
  | "dirty"
  | "validating"
  | "submitting"
  | "success"
  | "error";

import { FieldSchema, FormValues } from "../types/schema";

export const isFieldVisible = (
  field: FieldSchema,
  values: FormValues
): boolean => {
  if (!field.visibleWhen) return true;
  return values[field.visibleWhen.field] === field.visibleWhen.equals;
};

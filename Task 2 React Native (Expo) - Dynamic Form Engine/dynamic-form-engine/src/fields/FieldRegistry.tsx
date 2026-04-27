import React from "react";
import { FieldSchema, FieldValue, FormValues } from "../types/schema";

export type FieldComponentProps = {
  field: FieldSchema;
  value: FieldValue | undefined;
  values: FormValues;
  onChange: (value: FieldValue) => void;
  disabled?: boolean;
};

export type FieldComponent = React.ComponentType<FieldComponentProps>;
export type FieldRegistry = Record<string, FieldComponent>;

export const createFieldRegistry = (base: FieldRegistry = {}): FieldRegistry => ({ ...base });

export const registerFieldType = (
  registry: FieldRegistry,
  type: string,
  component: FieldComponent
): FieldRegistry => ({
  ...registry,
  [type]: component,
});

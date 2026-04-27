import {
  FieldSchema,
  FieldValue,
  FormValues,
  ValidationError,
  ValidationRule,
  ValidatorFn,
} from "../types/schema";
import { isFieldVisible } from "./visibility";

export type ValidatorRegistry = Record<string, ValidatorFn>;

const toStringValue = (value: FieldValue | undefined): string => {
  if (typeof value === "string") return value;
  if (typeof value === "number" || typeof value === "boolean") return String(value);
  return "";
};

const getRuleError = (
  rule: ValidationRule,
  field: FieldSchema,
  value: FieldValue | undefined,
  values: FormValues,
  validatorRegistry: ValidatorRegistry
): string | null => {
  const label = field.label;
  const stringValue = toStringValue(value);

  if (rule.type === "required") {
    const emptyArray = Array.isArray(value) && value.length === 0;
    const emptyValue = value === undefined || value === null || stringValue.trim() === "";
    if (emptyValue || emptyArray) return rule.message ?? `${label} is required`;
    return null;
  }

  if (rule.type === "minLength" && stringValue.length < rule.value) {
    return rule.message ?? `${label} must be at least ${rule.value} characters`;
  }

  if (rule.type === "maxLength" && stringValue.length > rule.value) {
    return rule.message ?? `${label} must be at most ${rule.value} characters`;
  }

  if (rule.type === "regex") {
    const matcher = new RegExp(rule.value, rule.flags);
    if (stringValue.length > 0 && !matcher.test(stringValue)) {
      return rule.message ?? `${label} has invalid format`;
    }
  }

  if (rule.type === "custom") {
    const candidate =
      typeof rule.validator === "string"
        ? validatorRegistry[rule.validator]
        : rule.validator;
    if (!candidate) return `Custom validator for ${label} is missing`;
    return candidate(value, values) ?? null;
  }

  return null;
};

export const validateField = (
  field: FieldSchema,
  values: FormValues,
  validatorRegistry: ValidatorRegistry = {}
): string[] => {
  if (!field.validation) return [];
  if (!isFieldVisible(field, values)) return [];

  const value = values[field.id];
  return field.validation
    .map((rule) => getRuleError(rule, field, value, values, validatorRegistry))
    .filter((item): item is string => item !== null);
};

export const validateForm = (
  fields: FieldSchema[],
  values: FormValues,
  validatorRegistry: ValidatorRegistry = {}
): ValidationError[] => {
  return fields.flatMap((field) =>
    validateField(field, values, validatorRegistry).map((message) => ({
      fieldId: field.id,
      message,
    }))
  );
};

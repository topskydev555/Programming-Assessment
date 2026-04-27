import React from "react";
import { Pressable, StyleSheet, Text, View } from "react-native";
import { isFieldVisible } from "../engine/visibility";
import { ValidatorRegistry } from "../engine/validation";
import {
  createFieldRegistry,
  FieldRegistry,
  registerFieldType,
} from "../fields/FieldRegistry";
import { defaultFieldRegistry } from "../fields/defaultFields";
import { useDynamicForm } from "../hooks/useDynamicForm";
import { FormSchema, FormValues } from "../types/schema";

type CustomFieldRegistration = {
  type: string;
  component: FieldRegistry[string];
};

type FormRendererProps = {
  schema: FormSchema;
  initialValues?: FormValues;
  customValidators?: ValidatorRegistry;
  customFields?: CustomFieldRegistration[];
  onSubmit: (values: FormValues) => Promise<unknown>;
};

export const FormRenderer = ({
  schema,
  initialValues,
  customValidators,
  customFields = [],
  onSubmit,
}: FormRendererProps) => {
  const registry = customFields.reduce(
    (acc, item) => registerFieldType(acc, item.type, item.component),
    createFieldRegistry(defaultFieldRegistry)
  );
  const { formState, setValue, submit } = useDynamicForm({
    schema,
    initialValues,
    validators: customValidators,
    onSubmit,
  });

  return (
    <View style={styles.container}>
      {!!schema.title && <Text style={styles.formTitle}>{schema.title}</Text>}
      {schema.fields
        .filter((field) => isFieldVisible(field, formState.values))
        .map((field) => {
          const FieldComponent = registry[field.type];
          if (!FieldComponent) {
            return (
              <View key={field.id}>
                <Text style={styles.label}>{field.label}</Text>
                <Text style={styles.error}>Unsupported field type: {field.type}</Text>
              </View>
            );
          }

          const errors = formState.errors
            .filter((error) => error.fieldId === field.id)
            .map((error) => error.message);
          return (
            <View key={field.id} style={styles.fieldBlock}>
              <Text style={styles.label}>{field.label}</Text>
              <FieldComponent
                field={field}
                value={formState.values[field.id]}
                values={formState.values}
                onChange={(nextValue) => setValue(field.id, nextValue)}
                disabled={formState.status === "submitting"}
              />
              {errors.map((message) => (
                <Text key={message} style={styles.error}>
                  {message}
                </Text>
              ))}
            </View>
          );
        })}

      <Pressable
        onPress={() => {
          void submit();
        }}
        disabled={formState.status === "submitting"}
        style={styles.submitButton}
      >
        <Text style={styles.submitLabel}>
          {formState.status === "submitting" ? "Submitting..." : "Submit"}
        </Text>
      </Pressable>

      <Text style={styles.status}>Status: {formState.status}</Text>
      <Text style={styles.valueDump}>
        Values: {JSON.stringify(formState.values, null, 2)}
      </Text>
    </View>
  );
};

const styles = StyleSheet.create({
  container: { gap: 12 },
  formTitle: { fontSize: 18, fontWeight: "700" },
  fieldBlock: { gap: 8 },
  label: { fontWeight: "600" },
  error: { color: "#b42318" },
  submitButton: {
    paddingHorizontal: 14,
    paddingVertical: 12,
    backgroundColor: "#2f7cff",
    borderRadius: 8,
    alignSelf: "flex-start",
  },
  submitLabel: { color: "#fff", fontWeight: "700" },
  status: { fontWeight: "700" },
  valueDump: { fontFamily: "monospace", fontSize: 12 },
});

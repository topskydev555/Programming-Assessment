import React from "react";
import { Pressable, StyleSheet, Text, TextInput, View } from "react-native";
import { FieldRegistry } from "./FieldRegistry";
import { PrimitiveFieldValue } from "../types/schema";

const TextField: FieldRegistry["text"] = ({ field, value, onChange, disabled }) => (
  <TextInput
    editable={!disabled}
    value={typeof value === "string" ? value : ""}
    onChangeText={(text) => onChange(text)}
    placeholder={field.placeholder}
    style={styles.input}
  />
);

const DateField: FieldRegistry["date"] = ({ field, value, onChange, disabled }) => (
  <TextInput
    editable={!disabled}
    value={typeof value === "string" ? value : ""}
    onChangeText={(text) => onChange(text)}
    placeholder={field.placeholder ?? "YYYY-MM-DD"}
    style={styles.input}
  />
);

const SelectButtons = ({
  selected,
  options,
  onSelect,
}: {
  selected: PrimitiveFieldValue | PrimitiveFieldValue[] | undefined;
  options: { label: string; value: PrimitiveFieldValue }[];
  onSelect: (value: PrimitiveFieldValue) => void;
}) => (
  <View style={styles.optionGroup}>
    {options.map((option) => {
      const isSelected = Array.isArray(selected)
        ? selected.includes(option.value)
        : selected === option.value;
      return (
        <Pressable
          key={`${option.value}`}
          onPress={() => onSelect(option.value)}
          style={[styles.optionButton, isSelected && styles.optionButtonSelected]}
        >
          <Text style={styles.optionLabel}>{option.label}</Text>
        </Pressable>
      );
    })}
  </View>
);

const SelectField: FieldRegistry["select"] = ({ field, value, onChange }) => (
  <SelectButtons
    selected={value as PrimitiveFieldValue | undefined}
    options={field.options ?? []}
    onSelect={(nextValue) => onChange(nextValue)}
  />
);

const MultiSelectField: FieldRegistry["multi-select"] = ({ field, value, onChange }) => {
  const selected = Array.isArray(value) ? value : [];
  return (
    <SelectButtons
      selected={selected}
      options={field.options ?? []}
      onSelect={(candidate) => {
        const next = selected.includes(candidate)
          ? selected.filter((item) => item !== candidate)
          : [...selected, candidate];
        onChange(next);
      }}
    />
  );
};

export const defaultFieldRegistry: FieldRegistry = {
  text: TextField,
  date: DateField,
  select: SelectField,
  "multi-select": MultiSelectField,
};

const styles = StyleSheet.create({
  input: {
    borderWidth: 1,
    borderColor: "#bbb",
    borderRadius: 8,
    paddingHorizontal: 12,
    paddingVertical: 10,
  },
  optionGroup: { gap: 8, flexDirection: "row", flexWrap: "wrap" },
  optionButton: {
    borderWidth: 1,
    borderColor: "#bbb",
    borderRadius: 999,
    paddingHorizontal: 12,
    paddingVertical: 8,
  },
  optionButtonSelected: {
    borderColor: "#2f7cff",
    backgroundColor: "#eaf2ff",
  },
  optionLabel: { fontWeight: "500" },
});

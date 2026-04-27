import { useMemo, useState } from "react";
import { FormController } from "../engine/formController";
import { ValidatorRegistry } from "../engine/validation";
import { FormSchema, FormValues } from "../types/schema";

type UseDynamicFormOptions = {
  schema: FormSchema;
  initialValues?: FormValues;
  validators?: ValidatorRegistry;
  onSubmit: (values: FormValues) => Promise<unknown>;
};

export const useDynamicForm = ({
  schema,
  initialValues,
  validators,
  onSubmit,
}: UseDynamicFormOptions) => {
  const controller = useMemo(
    () =>
      new FormController({
        schema,
        initialValues,
        validatorRegistry: validators,
      }),
    [schema, initialValues, validators]
  );

  const [formState, setFormState] = useState(controller.getState());

  const setValue = (fieldId: string, value: FormValues[string]) => {
    setFormState(controller.setFieldValue(fieldId, value));
  };

  const validate = () => setFormState(controller.validate());

  const submit = async () => {
    const nextState = await controller.submit(onSubmit);
    setFormState(nextState);
    return nextState;
  };

  return {
    formState,
    setValue,
    validate,
    submit,
  };
};

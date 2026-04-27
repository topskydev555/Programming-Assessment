import {
  FieldValue,
  FormSchema,
  FormStatus,
  FormValues,
  ValidationError,
} from "../types/schema";
import { ValidatorRegistry, validateForm } from "./validation";

export type FormControllerOptions = {
  schema: FormSchema;
  initialValues?: FormValues;
  validatorRegistry?: ValidatorRegistry;
};

export type FormSnapshot = {
  values: FormValues;
  errors: ValidationError[];
  status: FormStatus;
  dirtyFields: Set<string>;
};

export class FormController {
  private schema: FormSchema;
  private validatorRegistry: ValidatorRegistry;
  private snapshot: FormSnapshot;

  constructor(options: FormControllerOptions) {
    this.schema = options.schema;
    this.validatorRegistry = options.validatorRegistry ?? {};
    this.snapshot = {
      values: options.initialValues ?? {},
      errors: [],
      status: "pristine",
      dirtyFields: new Set<string>(),
    };
  }

  getState(): FormSnapshot {
    return {
      values: { ...this.snapshot.values },
      errors: [...this.snapshot.errors],
      status: this.snapshot.status,
      dirtyFields: new Set(this.snapshot.dirtyFields),
    };
  }

  setFieldValue(fieldId: string, value: FieldValue): FormSnapshot {
    this.snapshot.values = { ...this.snapshot.values, [fieldId]: value };
    this.snapshot.dirtyFields = new Set(this.snapshot.dirtyFields).add(fieldId);
    if (this.snapshot.status === "pristine") this.snapshot.status = "dirty";
    if (this.snapshot.status === "success" || this.snapshot.status === "error") {
      this.snapshot.status = "dirty";
    }
    return this.getState();
  }

  validate(): FormSnapshot {
    this.snapshot.status = "validating";
    this.snapshot.errors = validateForm(
      this.schema.fields,
      this.snapshot.values,
      this.validatorRegistry
    );
    this.snapshot.status = this.snapshot.errors.length > 0 ? "error" : "dirty";
    return this.getState();
  }

  async submit(submitter: (values: FormValues) => Promise<unknown>): Promise<FormSnapshot> {
    this.snapshot.status = "validating";
    this.snapshot.errors = validateForm(
      this.schema.fields,
      this.snapshot.values,
      this.validatorRegistry
    );
    if (this.snapshot.errors.length > 0) {
      this.snapshot.status = "error";
      return this.getState();
    }

    this.snapshot.status = "submitting";
    try {
      await submitter({ ...this.snapshot.values });
      this.snapshot.status = "success";
    } catch {
      this.snapshot.status = "error";
    }
    return this.getState();
  }
}

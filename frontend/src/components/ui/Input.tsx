import type { InputHTMLAttributes, ComponentType, ChangeEvent, ReactNode } from 'react';
// watercolor's Input supports value/onChange at runtime but the published types
// omit them; cast to keep the wrapper's typed API while passing props through.
import { Input as WcInputBase } from '@zeturn/watercolor-react';

const WcInput = WcInputBase as unknown as ComponentType<{
  label?: string;
  value?: string | number | readonly string[];
  onChange?: (e: ChangeEvent<HTMLInputElement>) => void;
  placeholder?: string;
  required?: boolean;
  error?: boolean;
  helperText?: ReactNode;
  size?: 'sm' | 'md' | 'lg';
  fullWidth?: boolean;
  className?: string;
  [key: string]: unknown;
}>;

interface InputProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'size' | 'color'> {
  label?: string;
  error?: string;
}

export default function Input({ label, error, className = '', ...props }: InputProps) {
  return (
    <WcInput
      label={label}
      error={!!error}
      helperText={error}
      size="md"
      fullWidth
      className={className}
      {...props}
    />
  );
}

import type { TextareaHTMLAttributes, ComponentType, ChangeEvent, ReactNode } from 'react';
import { TextField as WcTextFieldBase } from '@zeturn/watercolor-react';

const WcTextField = WcTextFieldBase as unknown as ComponentType<{
  label?: string;
  value?: string | number | readonly string[];
  onChange?: (e: ChangeEvent<HTMLTextAreaElement>) => void;
  placeholder?: string;
  required?: boolean;
  error?: ReactNode;
  multiline?: boolean;
  rows?: number;
  minRows?: number;
  size?: 'sm' | 'md' | 'lg';
  fullWidth?: boolean;
  className?: string;
  [key: string]: unknown;
}>;

interface TextareaProps extends Omit<TextareaHTMLAttributes<HTMLTextAreaElement>, 'size' | 'color'> {
  label?: string;
  error?: string;
}

export default function Textarea({ label, error, className = '', rows = 4, ...props }: TextareaProps) {
  return (
    <WcTextField
      label={label}
      error={error}
      multiline
      rows={rows}
      minRows={rows}
      size="md"
      fullWidth
      className={className}
      {...props}
    />
  );
}

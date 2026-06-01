import { useState } from 'react';
import { Copy, Check } from 'lucide-react';
import { Button } from './Button';
import './ui.css';

interface CodeBlockProps {
  code: string;
  language?: string;
}

export function CodeBlock({ code, language = 'log' }: CodeBlockProps) {
  const [copied, setCopied] = useState(false);

  const copy = async () => {
    await navigator.clipboard.writeText(code);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="ui-code">
      <div className="ui-code__header">
        <span>{language}</span>
        <Button variant="ghost" size="sm" onClick={copy}>
          {copied ? <Check size={14} /> : <Copy size={14} />}
        </Button>
      </div>
      <pre><code>{code}</code></pre>
    </div>
  );
}

interface MarkdownMessageProps {
  content: string;
}

function renderInline(text: string) {
  const parts = text.split(/(`[^`]+`|\*\*[^*]+\*\*)/g);
  return parts.map((part, index) => {
    if (part.startsWith('`') && part.endsWith('`')) {
      return <code key={index}>{part.slice(1, -1)}</code>;
    }
    if (part.startsWith('**') && part.endsWith('**')) {
      return <strong key={index}>{part.slice(2, -2)}</strong>;
    }
    return <span key={index}>{part}</span>;
  });
}

export function MarkdownMessage({ content }: MarkdownMessageProps) {
  const blocks = content.split('\n');
  return (
    <div className="markdown-message">
      {blocks.map((line, index) => {
        const trimmed = line.trim();
        if (!trimmed) return <br key={index} />;
        if (trimmed.startsWith('- ')) {
          return (
            <li key={index} style={{ marginLeft: 16 }}>
              {renderInline(trimmed.slice(2))}
            </li>
          );
        }
        if (trimmed.startsWith('### ')) {
          return <h4 key={index}>{renderInline(trimmed.slice(4))}</h4>;
        }
        if (trimmed.startsWith('## ')) {
          return <h3 key={index}>{renderInline(trimmed.slice(3))}</h3>;
        }
        return <p key={index}>{renderInline(trimmed)}</p>;
      })}
    </div>
  );
}

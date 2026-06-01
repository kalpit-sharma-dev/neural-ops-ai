import { useEffect, useState } from 'react';

interface TypingTextProps {
  text: string;
  speed?: number;
  className?: string;
}

export function TypingText({ text, speed = 18, className }: TypingTextProps) {
  const [visible, setVisible] = useState('');

  useEffect(() => {
    setVisible('');
    let i = 0;
    const timer = setInterval(() => {
      i += 1;
      setVisible(text.slice(0, i));
      if (i >= text.length) clearInterval(timer);
    }, speed);
    return () => clearInterval(timer);
  }, [text, speed]);

  return (
    <p className={className}>
      {visible}
      {visible.length < text.length && <span className="typing-cursor">|</span>}
    </p>
  );
}

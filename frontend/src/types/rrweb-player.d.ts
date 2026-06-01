declare module 'rrweb-player' {
  import type { eventWithTime } from 'rrweb';

  interface RRWebPlayerOptions {
    target: HTMLElement;
    props: {
      events: eventWithTime[];
      width?: number;
      height?: number;
      autoPlay?: boolean;
      showController?: boolean;
    };
  }

  export default class rrwebPlayer {
    constructor(options: RRWebPlayerOptions);
    destroy(): void;
  }
}

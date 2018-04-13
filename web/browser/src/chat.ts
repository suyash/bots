import Channel, { RecvChannel } from "./channel";
import { Message } from "./message";

export default class Chat {
    private channel: Channel<Message>;
    private socket: WebSocket;
    private openPromise: Promise<Event>;
    private closePromise: Promise<CloseEvent>;

    constructor(url: string) {
        this.channel = new Channel();
        this.socket = new WebSocket(url);

        this.openPromise = new Promise((resolve: (e: Event) => void): void => {
            this.socket.onopen = resolve;
        });

        this.closePromise = new Promise((resolve: (e: CloseEvent) => void): void => {
            this.socket.onclose = (e: CloseEvent): void => {
                this.channel.close(); // closing channel on connection close
                resolve(e);
            };
        });

        this.socket.onmessage = this.handleMessage;
    }

    public closed(): boolean {
        return this.channel.closed;
    }

    public open(): Promise<Event> {
        return this.openPromise;
    }

    public close(): Promise<Event> {
        return this.closePromise;
    }

    public say(text: string): void {
        const msg: Message = { source: "user", type: "message", text };
        return this.socket.send(JSON.stringify(msg));
    }

    public sayInThread(text: string, threadId: number): void {
        const msg: Message = { source: "user", type: "message", text, threadId };
        return this.socket.send(JSON.stringify(msg));
    }

    public messages(): RecvChannel<Message> {
        return this.channel;
    }

    private handleMessage = (e: MessageEvent): void => {
        const data: Blob = e.data;
        const reader: FileReader = new FileReader();
        reader.readAsText(data);
        reader.onloadend = (): void => {
            this.channel.send(JSON.parse(reader.result));
        };
    }
}

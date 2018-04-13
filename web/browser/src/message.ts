// The types here need to have 1:1 correspondence with message.go

export type ItemType = "message" | "thread" | "update";

export type ItemSource = "bot" | "user";

export interface Cursor {
    type: ItemType;
    id: number;
}

export interface Message {
    attachments?: Attachment[];
    source: ItemSource;
    type: ItemType;
    text: string;
    threadId?: number;
    replyId?: number;
    id?: number;
    prev?: Cursor;
    next?: Cursor;
}

export interface Attachment {
    type: "image" | "audio" | "video" | "location" | "file_download";
}

export interface Image extends Attachment {
    url: string;
    title: string;
    text: string;
    alt: string;
}

export interface Audio extends Attachment {
    url: string;
    title: string;
    text: string;
}

export interface Video extends Attachment {
    url: string;
    title: string;
    text: string;
}

export interface Location extends Attachment {
    lat: number;
    long: number;
    title: string;
    text: string;
}

export interface FileDownload extends Attachment {
    url: string;
    text: string;
}

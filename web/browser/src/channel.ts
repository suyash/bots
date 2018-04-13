/**
 * https://github.com/tj/channel.js
 *
 * implemented with interfaces, and without the deferred dependency.
 */

export const errSendOnClosed: Error =  new Error("send on closed channel");
export const errChanClosed: Error =  new Error("channel already closed");

interface Deferred<T> {
    promise: Promise<T>;
    resolve(t?: T): void;
    reject(err: Error): void;
}

function newDeferred<T>(): Deferred<T> {
    const ans: Deferred<T> = {
        promise: null,
        reject: null,
        resolve: null,
    };

    // the question mark is needed because for void promises it'll tell you to pass a parameter
    // but passing anything will also be an error, + for void promises you should be able to call
    // the function without any parameters.
    ans.promise = new Promise((resolve: (t?: T) => void, reject: (err: Error) => void): void => {
        ans.resolve = resolve;
        ans.reject = reject;
    });

    return ans;
}

/**
 * RecvChannel supports async iteration protocol.
 * recv() is simply be an alias for next().
 */
export interface RecvChannel<T> extends AsyncIterable<T> {
    /**
     * Can't make recv() simply return value (with zero value when closed)
     * breaks AsyncIterable behavior when all open receives are closed in close().
     */

    // NOTE: this should not be needed, but programs importing do not compile without this
    [Symbol.asyncIterator](): AsyncIterator<T>;

    recv(): Promise<{value: T, done: boolean}>;
    next(): Promise<{value: T, done: boolean}>;
}

export interface SendChannel<T> {
    send(value: T): Promise<void>;
}

export default class Channel<T> implements SendChannel<T>, RecvChannel<T> {
    public closed: boolean;

    private capacity: number;
    private values: T[];
    private sends: Array<{value: T, deferred: Deferred<void>}>;
    private recvs: Array<Deferred<{value: T, done: boolean}>>;

    /**
     * Initialize channel with the given buffer `capacity`. By default
     * the channel is unbuffered. A channel is basically a FIFO queue
     * for use with async/await or co().
     */

    constructor(capacity: number = 0) {
        this.capacity = capacity;
        this.values = [];
        this.sends = [];
        this.recvs = [];
        this.closed = false;
    }

    public [Symbol.asyncIterator] = (): AsyncIterator<T> => {
        return this;
    }

    /**
     * Send value, blocking unless there is room in the buffer.
     *
     * Calls to send() on a closed buffer will error.
     */

    public send(value: T): Promise<void> {
        if (this.closed) {
            return Promise.reject(errSendOnClosed);
        }

        // recv pending
        if (this.recvs.length) {
            this.recvs.shift().resolve({ value, done: false });
            return Promise.resolve();
        }

        // room in buffer
        if (this.values.length < this.capacity) {
            this.values.push(value);
            return Promise.resolve();
        }

        // no recv pending, block
        const deferred: Deferred<void> = newDeferred();
        this.sends.push({ value, deferred });
        return deferred.promise;
    }

    /**
     * Receive returns a value or blocks until one is present.
     *
     * A recv() on a closed channel will return undefined.
     */

    public recv(): Promise<{value: T, done: boolean}> {
        // values in buffer
        if (this.values.length) {
            return Promise.resolve({value: this.values.shift(), done: false});
        }

        // unblock pending sends
        if (this.sends.length) {
            const send: { value: T; deferred: Deferred<void>; } = this.sends.shift();

            if (this.closed) {
                send.deferred.reject(errSendOnClosed);
                // receive on closed channel gets zero value
                // but sending null because not figuring out zero value
                return Promise.resolve({value: null, done: true});
            }

            send.deferred.resolve();
            return Promise.resolve({value: send.value, done: false});
        }

        // closed
        if (this.closed) {
            // receive on closed channel gets zero value
            // but sending null because not figuring out zero value
            return Promise.resolve({value: null, done: true});
        }

        // no values, block
        const promise: Deferred<{value: T, done: boolean}> = newDeferred();
        this.recvs.push(promise);
        return promise.promise;
    }

    /**
     * next is an alias for recv to satisfy AsyncIterator
     */

    public next(): Promise<{value: T, done: boolean}> {
        return this.recv();
    }

    /**
     * Close the channel. Any pending recv() calls will be unblocked.
     *
     * Subsequent close() calls will throw.
     */

    public close(): void {
        if (this.closed) {
            throw errChanClosed;
        }

        this.closed = true;
        const recvs: Array<Deferred<{value: T, done: boolean}>> = this.recvs;
        this.recvs = [];

        for (const p of recvs) {
            p.resolve({value: null, done: true});
        }
    }
}

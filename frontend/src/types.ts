export type Topic = {
    id: number;
    name: string;
};

export type Post = {
    id: number;
    topicID: number;
    title: string;
    content: string;
    author?: string;
};

export type Comment = {
    id: number;
    postID: number;
    content: string;
    author?: string;
};

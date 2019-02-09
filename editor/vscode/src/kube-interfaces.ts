
export interface Metadata {
    name: string;
    namespace: string;
};

export interface PodSpec {
    containers: Container[];
    nodeName: string;
}

export interface Pod {
    metadata: Metadata;
    spec: PodSpec;
}

export interface PodList {
    items: Pod[];
}

export interface Container {
    name: string;
    image: string;
}

export interface ServiceSpec {
    selector: { [key: string]: string; };
}

export interface Service {
    metadata: Metadata;
    spec: ServiceSpec;
}

export interface ServiceList {
    items: Service[];
}

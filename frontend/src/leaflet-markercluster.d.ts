import 'leaflet';

declare module 'leaflet' {
    interface MarkerClusterGroupOptions {
        maxClusterRadius?: number;
        disableClusteringAtZoom?: number;
    }

    class MarkerClusterGroup extends FeatureGroup {
        clearLayers(): this;
        addLayer(layer: Layer): this;
        getBounds(): LatLngBounds;
    }

    function markerClusterGroup(options?: MarkerClusterGroupOptions): MarkerClusterGroup;
}

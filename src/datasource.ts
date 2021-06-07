import {
  DataSourceInstanceSettings,
  DataQueryRequest,
  DataQueryResponse,
  PanelPlugin,
  DataQuery,
  DataSourceJsonData,
  DataSourceApi,
  DataSourcePluginMeta,
  KeyValue,
} from '@grafana/data';
import { DataSourceWithBackend, getDataSourceSrv, HealthCheckResult } from '@grafana/runtime';
import { ObjectDataSourceOptions, ObjectQuery, ObjectPanelType } from './types';
import { ObjectPanel } from './Components/Panel';
import { mergeAll } from 'rxjs/operators';
import { Observable, from } from 'rxjs';

interface QueryMapEntry<
  TQuery extends DataQuery = DataQuery,
  TOptions extends DataSourceJsonData = DataSourceJsonData
> {
  datasource: DataSourceApi<TQuery, TOptions>;
  request: DataQueryRequest<TQuery>;
}

export class DataSource extends DataSourceWithBackend<ObjectQuery, ObjectDataSourceOptions> {
  panel: PanelPlugin<ObjectPanelType>;
  constructor(public instanceSettings: DataSourceInstanceSettings<ObjectDataSourceOptions>) {
    super(instanceSettings);
    this.panel = new PanelPlugin<ObjectPanelType>(ObjectPanel).setPanelOptions((builder) => {});
  }

  query(request: DataQueryRequest<ObjectQuery>): Observable<DataQueryResponse> {
    const dsPromises = request.targets.map((target: ObjectQuery) => getDataSourceSrv().get(target.config.uid));

    const resultPromise = Promise.all(dsPromises)
      .then((dataSources) => {
        return dataSources.reduce(
          (
            mapping: KeyValue<QueryMapEntry<DataQuery, ObjectDataSourceOptions>>,
            curr: DataSourceApi<DataQuery, DataSourceJsonData>
          ) => {
            if (!mapping[curr.uid]) {
              const targets: ObjectQuery[] = [];
              mapping[curr.uid] = {
                datasource: curr,
                request: {
                  ...request,
                  targets: targets,
                },
              };
            }

            return mapping;
          },
          {}
        );
      })
      .then((mapping) => {
        request.targets.forEach((target) => {
          if (!target.hide && target.config.uid && target.config.query) {
            //if (isBackendPlugin(mapping[target.config.uid].datasource.meta)) {
            //  mapping[target.config.uid].request.targets.push(target);
            //} else {
            mapping[target.config.uid].request.targets.push(target.config.query);
            //}
          }
        });
        return mapping;
      })
      .then((populatedMapping) => {
        const results: Array<Promise<DataQueryResponse>> = [];
        Object.values(populatedMapping).forEach((entry) => {
          // if (isObjectQueryRequest(entry.request)) {
          //   results.push(super.query(entry.request).toPromise());
          // } else {
          if (entry.datasource) {
            const result = entry.datasource.query(entry.request);
            if (isPromise(result)) {
              results.push(result);
            } else {
              results.push(result.toPromise());
            }
          }
          //}
        });
        return results;
      })
      .then((results) => Promise.all(results));

    return from(resultPromise).pipe(mergeAll());
  }

  callHealthCheck(): Promise<HealthCheckResult> {
    console.log('calling callHealthCheck');
    return super.callHealthCheck();
  }
}

export const isPromise = (
  response: Promise<DataQueryResponse> | Observable<DataQueryResponse>
): response is Promise<DataQueryResponse> => {
  return (response as Promise<DataQueryResponse>).then !== undefined;
};

export const isObjectQueryRequest = (
  request: DataQueryRequest<ObjectQuery> | DataQueryRequest<DataQuery>
): request is DataQueryRequest<ObjectQuery> => {
  for (let target of request.targets) {
    if (!(target as ObjectQuery).config) {
      return false;
    }
  }
  return true;
};

export const isObservable = (
  response: Promise<DataQueryResponse> | Observable<DataQueryResponse>
): response is Observable<DataQueryResponse> => {
  return (response as Observable<DataQueryResponse>).subscribe !== undefined;
};

export const isBackendPlugin = (pluginMeta: DataSourcePluginMeta): boolean => {
  return (pluginMeta as any)['backend'] !== undefined || (pluginMeta as any)['signature']! === 'internal';
};

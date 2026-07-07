import React, { useCallback, useEffect, useRef, useState } from 'react';
import * as echarts from 'echarts';
import { Province } from '../types';

interface Props {
  provinces: Province[];
  onProvinceClick: (name: string) => void;
}

const GEO_URL = './china_map.json';

const gpName: Record<string, string> = {
  "上海市":"上海市","北京市":"北京市","天津市":"天津市","重庆市":"重庆市",
  "安徽省":"安徽省","福建省":"福建省","甘肃省":"甘肃省","广东省":"广东省",
  "贵州省":"贵州省","海南省":"海南省","河北省":"河北省","河南省":"河南省",
  "黑龙江省":"黑龙江省","湖北省":"湖北省","湖南省":"湖南省","吉林省":"吉林省",
  "江苏省":"江苏省","江西省":"江西省","辽宁省":"辽宁省","青海省":"青海省",
  "山东省":"山东省","山西省":"山西省","陕西省":"陕西省","四川省":"四川省",
  "云南省":"云南省","浙江省":"浙江省",
  "内蒙古自治区":"内蒙古自治区","宁夏回族自治区":"宁夏回族自治区",
  "广西壮族自治区":"广西壮族自治区","新疆维吾尔自治区":"新疆维吾尔自治区",
  "西藏自治区":"西藏自治区","台湾省":"台湾省",
  "香港特别行政区":"香港特别行政区","澳门特别行政区":"澳门特别行政区",
};
function echartsProvinceName(name: string): string {
  return gpName[name] || name;
}

const ChinaMap: React.FC<Props> = ({ provinces, onProvinceClick }) => {
  const chartRef = useRef<HTMLDivElement>(null);
  const [geoLoaded, setGeoLoaded] = useState(false);
  const chartInstance = useRef<echarts.ECharts | null>(null);

  useEffect(() => {
    fetch(GEO_URL).then(r => r.json()).then(geo => {
      echarts.registerMap('china', geo as any);
      setGeoLoaded(true);
    });
  }, []);

  const initChart = useCallback(() => {
    if (!geoLoaded || !chartRef.current) return;
    if (!chartInstance.current) chartInstance.current = echarts.init(chartRef.current);
    const maxCount = Math.max(...provinces.map(p => p.trial_count), 1);
    const data = provinces.map(p => ({ name: echartsProvinceName(p.name), value: p.trial_count }));
    chartInstance.current.setOption({
      tooltip: { trigger: 'item', formatter: '{b}: {c} 项试验' },
      visualMap: { min: 0, max: maxCount, text: ['多', '少'], inRange: { color: ['#e0f2fe', '#0369a1'] }, left: 'left', bottom: 20 },
      series: [{
        name: '临床试验分布', type: 'map', map: 'china', roam: true,
        zoom: window.innerWidth < 640 ? 1.3 : window.innerWidth < 1024 ? 1.15 : 1.05,
        selectedMode: false, label: { show: true, fontSize: Math.min(10, Math.max(7, window.innerWidth / 80)) },
        emphasis: { label: { color: '#fff' }, itemStyle: { areaColor: '#0c4a6e' } },
        data, itemStyle: { borderColor: '#fff', borderWidth: 1 },
      }],
    });
    chartInstance.current.off('click');
    chartInstance.current.on('click', (params: any) => {
      const province = provinces.find(p => echartsProvinceName(p.name) === params.name);
      if (province) onProvinceClick(province.name);
    });
  }, [geoLoaded, provinces, onProvinceClick]);
  useEffect(initChart, [initChart]);

  const [mapHeight, setMapHeight] = useState(600);
  useEffect(() => {
    const update = () => {
      const w = window.innerWidth;
      const h = window.innerHeight;
      setMapHeight(w < 640 ? Math.min(500, h * 0.65) : w < 1024 ? 500 : 600);
      chartInstance.current?.resize();
    };
    update();
    window.addEventListener('resize', update);
    return () => window.removeEventListener('resize', update);
  }, []);
  return <div ref={chartRef} style={{ width: '100%', height: mapHeight }} />;
};

export default ChinaMap;

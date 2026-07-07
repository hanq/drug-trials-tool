import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Search, Database } from 'lucide-react';

interface Props { onSearch: (q: string) => void; }

const Header: React.FC<Props> = ({ onSearch }) => {
  const [sq, setSq] = useState('');
  const nav = useNavigate();
  return (
    <header className="header">
      <div className="header-inner">
        <div className="header-left">
           <span className="header-logo" onClick={() => nav('/')}>
           <Database size={22} /> 就近找临床
           </span>
        </div>
        <div className="header-search">
          <input type="text" placeholder="搜索试验（标题/药物/适应症/登记号）" value={sq}
          onChange={e => setSq(e.target.value)}
          onKeyDown={e => { if (e.key === 'Enter') { onSearch(sq); nav('/?q=' + encodeURIComponent(sq)); } }} />
          <Search size={18} className="search-icon" onClick={() => { onSearch(sq); nav('/?q=' + encodeURIComponent(sq)); }} />
        </div>
        <div className="header-right"></div>
      </div>
    </header>
  );
};

export default Header;

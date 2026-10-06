import { Button } from "@/components/ui/button";
import { ArrowUpRight } from "lucide-react";

const ButtonWithIconDemo = () => {
  return (
    <Button className="relative text-sm font-medium rounded-full h-10 p-1 ps-6 pe-14 group transition-all duration-500 hover:ps-14 hover:pe-6 w-fit overflow-hidden cursor-pointer">
      <span className="relative z-10 transition-all duration-500">
        Войти
      </span>
      <div className="absolute right-1 w-8 h-8 bg-(--itlib-frost-glow) text-foreground rounded-full flex items-center justify-center transition-all duration-500 group-hover:right-[calc(100%-44px)] group-hover:rotate-45">
        <ArrowUpRight size={16} className="text-black" />
      </div>
    </Button>
  );
};

export default ButtonWithIconDemo;


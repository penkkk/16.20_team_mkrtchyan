import { Button } from "@/components/ui/button";
import { ArrowUpRight } from "lucide-react";

const ButtonWithIconDemo = () => {
  return (
    <Button className="h-10 gap-2 px-4 text-sm font-medium">
      Войти
      <ArrowUpRight size={16} aria-hidden="true" />
    </Button>
  );
};

export default ButtonWithIconDemo;

